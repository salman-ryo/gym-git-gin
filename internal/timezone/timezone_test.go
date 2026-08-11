package timezone_test

import (
	"gymgit/backend/internal/timezone"
	"testing"
)

func TestValidateIANA(t *testing.T) {
	validTimezones := []string{
		"UTC",
		"America/New_York",
		"America/Los_Angeles",
		"Europe/London",
		"Asia/Kolkata",
		"Asia/Tokyo",
	}

	for _, tz := range validTimezones {
		if err := timezone.ValidateIANA(tz); err != nil {
			t.Errorf("expected timezone '%s' to be valid, got error: %v", tz, err)
		}
	}

	invalidTimezones := []string{
		"",
		"   ",
		"Invalid/Timezone_Name_12345",
		"GMT+25",
	}

	for _, tz := range invalidTimezones {
		if err := timezone.ValidateIANA(tz); err == nil {
			t.Errorf("expected timezone '%s' to be invalid, but got no error", tz)
		}
	}
}

func TestLoadLocationFallback(t *testing.T) {
	loc := timezone.LoadLocation("Invalid/TZ")
	if loc.String() != "UTC" {
		t.Errorf("expected fallback location 'UTC', got '%s'", loc.String())
	}

	locNY := timezone.LoadLocation("America/New_York")
	if locNY.String() != "America/New_York" {
		t.Errorf("expected location 'America/New_York', got '%s'", locNY.String())
	}
}

func TestGetUserToday(t *testing.T) {
	loc := timezone.LoadLocation("Asia/Kolkata")
	today := timezone.GetUserToday(loc)

	if today.Hour() != 0 || today.Minute() != 0 || today.Second() != 0 || today.Nanosecond() != 0 {
		t.Errorf("expected GetUserToday to be 00:00:00.000, got %v", today)
	}

	dateStr := timezone.FormatUserToday(loc)
	if len(dateStr) != 10 {
		t.Errorf("expected formatted date YYYY-MM-DD, got '%s'", dateStr)
	}
}

func TestDateSegregationHelpers(t *testing.T) {
	loc := timezone.LoadLocation("America/Los_Angeles")
	todayStr := timezone.FormatUserToday(loc)

	if !timezone.IsTodayDate(todayStr, loc) {
		t.Errorf("expected IsTodayDate to return true for '%s'", todayStr)
	}

	yesterdayTime := timezone.GetUserToday(loc).AddDate(0, 0, -1)
	yesterdayStr := yesterdayTime.Format("2006-01-02")

	if !timezone.IsHistoricalDate(yesterdayStr, loc) {
		t.Errorf("expected IsHistoricalDate to return true for '%s'", yesterdayStr)
	}

	if timezone.IsHistoricalDate(todayStr, loc) {
		t.Errorf("expected IsHistoricalDate to return false for today '%s'", todayStr)
	}
}
