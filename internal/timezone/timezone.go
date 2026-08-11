package timezone

import (
	"fmt"
	"strings"
	"time"
)

// DefaultTimezone is the fallback IANA timezone
const DefaultTimezone = "UTC"

// ValidateIANA validates whether a given timezone string is a valid IANA location
func ValidateIANA(tz string) error {
	tz = strings.TrimSpace(tz)
	if tz == "" {
		return fmt.Errorf("timezone string cannot be empty")
	}
	_, err := time.LoadLocation(tz)
	if err != nil {
		return fmt.Errorf("invalid IANA timezone '%s': %w", tz, err)
	}
	return nil
}

// LoadLocation safely loads a time.Location for a given timezone string, defaulting to UTC on error
func LoadLocation(tz string) *time.Location {
	tz = strings.TrimSpace(tz)
	if tz == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.UTC
	}
	return loc
}

// GetUserToday returns the normalized start of the current day (00:00:00) in the user's localized wall-clock timezone
func GetUserToday(loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
}

// FormatUserToday returns YYYY-MM-DD string of the user's current wall-clock today
func FormatUserToday(loc *time.Location) string {
	return GetUserToday(loc).Format("2006-01-02")
}

// IsHistoricalDate checks if a target date string (YYYY-MM-DD) strictly precedes user's localized today date string
func IsHistoricalDate(targetDateStr string, loc *time.Location) bool {
	todayStr := FormatUserToday(loc)
	return targetDateStr < todayStr
}

// IsTodayDate checks if a target date string (YYYY-MM-DD) matches user's localized today date string
func IsTodayDate(targetDateStr string, loc *time.Location) bool {
	todayStr := FormatUserToday(loc)
	return targetDateStr == todayStr
}
