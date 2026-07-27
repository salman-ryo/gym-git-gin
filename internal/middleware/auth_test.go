package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestAuthMiddleware_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware("secret"))
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_ValidBearerHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-supabase-secret"
	testUUID := uuid.New()

	claims := SupabaseClaims{
		Email: "gymnerd@example.com",
		UserMetadata: map[string]interface{}{
			"full_name":  "Gym Bro",
			"avatar_url": "https://example.com/avatar.png",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   testUUID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))

	router := gin.New()
	router.Use(AuthMiddleware(secret))
	var capturedUserID uuid.UUID
	router.GET("/protected", func(c *gin.Context) {
		val, _ := c.Get(ContextAuthUserIDKey)
		capturedUserID = val.(uuid.UUID)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if capturedUserID != testUUID {
		t.Errorf("expected captured UUID %s, got %s", testUUID, capturedUserID)
	}
}

func TestAuthMiddleware_CookieIgnored(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-supabase-secret"
	testUUID := uuid.New()

	claims := SupabaseClaims{
		Email: "cookie@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   testUUID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))

	router := gin.New()
	router.Use(AuthMiddleware(secret))
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Request with ONLY cookie (no Authorization header) must return 401 Unauthorized
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "sb-access-token", Value: tokenString})
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d for cookie-only request, got %d", http.StatusUnauthorized, w.Code)
	}
}
