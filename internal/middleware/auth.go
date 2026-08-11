package middleware

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"gymgit/backend/internal/models"
	"gymgit/backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	ContextAuthUserIDKey     = "auth_user_id"
	ContextUserEmailKey      = "user_email"
	ContextUserNameKey       = "user_name"
	ContextAvatarURLKey      = "user_avatar_url"
	ContextProviderKey       = "user_provider"
	ContextResolvedUserIDKey = "resolved_user_id"
)

// SupabaseClaims represents the payload structure of a Supabase Auth JWT
type SupabaseClaims struct {
	Email        string                 `json:"email"`
	UserMetadata map[string]interface{} `json:"user_metadata"`
	AppMetadata  map[string]interface{} `json:"app_metadata"`
	jwt.RegisteredClaims
}

var (
	// jwksCache holds interface{} so it can store either *rsa.PublicKey or *ecdsa.PublicKey
	jwksCache = make(map[string]interface{})
	jwksMutex sync.RWMutex
)

// AuthMiddleware creates a Gin middleware that extracts and validates a Supabase JWT
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	if jwtSecret == "" {
		log.Println("⚠️ [AuthMiddleware WARNING] SUPABASE_JWT_SECRET environment variable is empty!")
	}
	return func(c *gin.Context) {
		tokenString := extractToken(c)

		if tokenString == "" {
			models.SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authentication token in header", nil)
			c.Abort()
			return
		}

		claims := &SupabaseClaims{}
		token, err := parseSupabaseJWT(tokenString, jwtSecret, claims)

		if err != nil || token == nil || !token.Valid {
			errMsg := "Invalid or expired authentication token"
			if err != nil {
				errMsg = err.Error()
			}
			log.Printf("[AuthMiddleware 401] Token validation failed: %v", errMsg)
			models.SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired authentication token", []string{errMsg})
			c.Abort()
			return
		}

		// Parse sub claim to UUID
		authUserID, err := uuid.Parse(claims.Subject)
		if err != nil {
			models.SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID claim in token", nil)
			c.Abort()
			return
		}

		// Extract metadata defaults
		var name, avatarURL string
		provider := "email"

		if claims.UserMetadata != nil {
			if n, ok := claims.UserMetadata["full_name"].(string); ok && n != "" {
				name = n
			} else if n, ok := claims.UserMetadata["name"].(string); ok {
				name = n
			}
			if a, ok := claims.UserMetadata["avatar_url"].(string); ok {
				avatarURL = a
			}
		}

		if claims.AppMetadata != nil {
			if p, ok := claims.AppMetadata["provider"].(string); ok && p != "" {
				provider = p
			}
		}

		// Store in Gin Request Context
		c.Set(ContextAuthUserIDKey, authUserID)
		c.Set(ContextUserEmailKey, claims.Email)
		c.Set(ContextUserNameKey, name)
		c.Set(ContextAvatarURLKey, avatarURL)
		c.Set(ContextProviderKey, provider)

		c.Next()
	}
}

// extractToken retrieves the JWT token string strictly from the Authorization: Bearer <token> header
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		token := strings.TrimSpace(parts[1])
		if token != "" && token != "null" && token != "undefined" {
			// Strip any accidental quotes
			return strings.Trim(token, `"`)
		}
	}

	return ""
}

// parseSupabaseJWT dynamically parses HS256, RS256, or ES256 Supabase tokens
func parseSupabaseJWT(tokenString, jwtSecret string, claims *SupabaseClaims) (*jwt.Token, error) {
	parser := jwt.NewParser()
	unverifiedToken, _, err := parser.ParseUnverified(tokenString, &SupabaseClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token structure: %w", err)
	}

	alg, _ := unverifiedToken.Header["alg"].(string)

	opts := []jwt.ParserOption{
		jwt.WithLeeway(5 * time.Minute),
		jwt.WithAudience("authenticated"),
	}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		switch alg {
		case "RS256", "ES256":
			kid, _ := token.Header["kid"].(string)
			return getJWKSPublicKey(kid)
		case "HS256":
			return getHMACKey(jwtSecret)
		default:
			return nil, fmt.Errorf("unsupported signing algorithm: %s", alg)
		}
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc, opts...)
	if err != nil {
		return nil, fmt.Errorf("signature validation failed (%s): %w", alg, err)
	}

	return token, nil
}

func getHMACKey(secret string) ([]byte, error) {
	if decoded, err := tryDecodeBase64Bytes(secret); err == nil && len(decoded) > 0 {
		return decoded, nil
	}
	return []byte(secret), nil
}

// getJWKSPublicKey fetches and caches JWKS public keys, supporting both RSA and ECDSA
func getJWKSPublicKey(kid string) (interface{}, error) {
	jwksMutex.RLock()
	pubKey, exists := jwksCache[kid]

	if !exists && kid == "" && len(jwksCache) > 0 {
		for _, k := range jwksCache {
			pubKey = k
			exists = true
			break
		}
	}
	jwksMutex.RUnlock()

	if exists {
		return pubKey, nil
	}

	jwksURL := os.Getenv("SUPABASE_JWKS_URL")
	if jwksURL == "" {
		return nil, fmt.Errorf("SUPABASE_JWKS_URL environment variable is missing for asymmetric token validation")
	}

	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			// RSA fields
			N string `json:"n"`
			E string `json:"e"`
			// ECDSA fields
			Crv string `json:"crv"`
			X   string `json:"x"`
			Y   string `json:"y"`
		} `json:"keys"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS payload: %w", err)
	}

	jwksMutex.Lock()
	defer jwksMutex.Unlock()

	var matchedKey interface{}

	for _, key := range jwks.Keys {
		var parsedKey interface{}

		if key.Kty == "RSA" {
			nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
			if err != nil {
				continue
			}
			eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
			if err != nil {
				continue
			}

			n := new(big.Int).SetBytes(nBytes)
			e := 0
			for _, b := range eBytes {
				e = (e << 8) | int(b)
			}
			parsedKey = &rsa.PublicKey{N: n, E: e}

		} else if key.Kty == "EC" {
			xBytes, errX := base64.RawURLEncoding.DecodeString(key.X)
			yBytes, errY := base64.RawURLEncoding.DecodeString(key.Y)
			if errX != nil || errY != nil {
				continue
			}

			x := new(big.Int).SetBytes(xBytes)
			y := new(big.Int).SetBytes(yBytes)

			// ES256 uses the P-256 curve
			parsedKey = &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
		}

		if parsedKey != nil {
			jwksCache[key.Kid] = parsedKey
			if key.Kid == kid || kid == "" {
				matchedKey = parsedKey
			}
		}
	}

	if matchedKey != nil {
		return matchedKey, nil
	}

	return nil, fmt.Errorf("no matching key found in JWKS for kid: %s", kid)
}

func tryDecodeBase64Bytes(s string) ([]byte, error) {
	encoders := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}
	for _, enc := range encoders {
		if decoded, err := enc.DecodeString(s); err == nil && len(decoded) > 0 {
			return decoded, nil
		}
	}
	return nil, fmt.Errorf("invalid base64")
}

// ResolveUserMiddleware resolves the database user profile ID from the auth_user_id
func ResolveUserMiddleware(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authUserIDVal, exists := c.Get(ContextAuthUserIDKey)
		if !exists {
			c.Next()
			return
		}
		authUserID := authUserIDVal.(uuid.UUID)

		if userRepo == nil {
			c.Next()
			return
		}

		user, err := userRepo.GetByAuthUserID(c.Request.Context(), authUserID)
		if err != nil || user == nil {
			// Profile not found - we don't abort because bootstrap needs to run
			c.Next()
			return
		}

		c.Set(ContextResolvedUserIDKey, user.ID)
		c.Next()
	}
}

// GetResolvedUserID retrieves the database primary key ID of the user from the context
func GetResolvedUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(ContextResolvedUserIDKey)
	if !exists {
		models.SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User profile not bootstrapped", nil)
		return uuid.Nil, false
	}
	return val.(uuid.UUID), true
}
