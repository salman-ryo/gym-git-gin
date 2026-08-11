package middleware

import (
	"time"

	"gymgit/backend/internal/timezone"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserLocationKey = "user_location"
	ContextUserTimezoneKey = "user_timezone"
)

// TimezoneMiddleware extracts timezone from X-Timezone header or timezone cookie and sets user_location in context
func TimezoneMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tz := c.GetHeader("X-Timezone")
		if tz == "" {
			if cookieTZ, err := c.Cookie("timezone"); err == nil && cookieTZ != "" {
				tz = cookieTZ
			}
		}

		if tz == "" {
			tz = timezone.DefaultTimezone
		}

		loc := timezone.LoadLocation(tz)

		c.Set(ContextUserTimezoneKey, loc.String())
		c.Set(ContextUserLocationKey, loc)

		c.Next()
	}
}

// GetUserLocationFromContext retrieves the user's *time.Location from Gin context, defaulting to UTC
func GetUserLocationFromContext(c *gin.Context) *time.Location {
	if locVal, exists := c.Get(ContextUserLocationKey); exists {
		if loc, ok := locVal.(*time.Location); ok && loc != nil {
			return loc
		}
	}
	return time.UTC
}
