package service

import (
	"testing"
	"time"

	"gymgit/backend/internal/models"
)

func TestCalculateTargetDaysPerWeek(t *testing.T) {
	tests := []struct {
		name       string
		categories []string
		expected   int
	}{
		{"PPL 6 day split", []string{"Push", "Pull", "Legs", "Push", "Pull", "Legs"}, 6},
		{"Upper Lower 4 day split", []string{"Upper", "Lower", "Upper", "Lower"}, 4},
		{"Full Body 3 day split", []string{"Full Body", "Full Body", "Full Body"}, 3},
		{"Minimal 2 categories", []string{"Upper", "Lower"}, 3}, // max(3, 2) = 3
		{"Empty categories", []string{}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateTargetDaysPerWeek(tt.categories)
			if got != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}

func TestMapPowerScoreToAnimeTier(t *testing.T) {
	tests := []struct {
		score        int
		expectedTier string
		expectedChar string
	}{
		{5, "D", "Aqua"},
		{14, "D", "Aqua"},
		{15, "C", "Mumen Rider"},
		{34, "C", "Mumen Rider"},
		{35, "B", "Tanjiro Kamado"},
		{55, "A", "Izuku Midoriya"},
		{70, "S", "Monkey D. Luffy"},
		{85, "S+", "Satoru Gojo"},
		{95, "SS", "Saitama"},
		{100, "SS", "Saitama"},
	}

	for _, tt := range tests {
		tier := MapPowerScoreToAnimeTier(tt.score)
		if tier.Tier != tt.expectedTier {
			t.Errorf("score %d: expected tier %s, got %s", tt.score, tt.expectedTier, tier.Tier)
		}
		if tier.Character != tt.expectedChar {
			t.Errorf("score %d: expected char %s, got %s", tt.score, tt.expectedChar, tier.Character)
		}
	}
}

func TestCalculatePowerScore(t *testing.T) {
	today := time.Now().UTC()
	logs := []models.GymLog{}

	// Generate 20 days of 1.25h "Push", "Pull", "Legs" workouts
	for i := 0; i < 20; i++ {
		dateStr := today.AddDate(0, 0, -i).Format("2006-01-02")
		wType := "Push"
		if i%3 == 1 {
			wType = "Pull"
		} else if i%3 == 2 {
			wType = "Legs"
		}
		logs = append(logs, models.GymLog{
			Date:        dateStr,
			Hours:       1.25,
			WorkoutType: wType,
		})
	}

	breakdown := CalculatePowerScore(logs, 4, 30, today)
	if breakdown.TotalScore <= 0 || breakdown.TotalScore > 100 {
		t.Errorf("expected score between 1 and 100, got %d", breakdown.TotalScore)
	}
	if breakdown.Consistency <= 0 {
		t.Errorf("expected consistency > 0, got %d", breakdown.Consistency)
	}
	if breakdown.DurationQuality <= 0 {
		t.Errorf("expected duration quality > 0, got %d", breakdown.DurationQuality)
	}
	if breakdown.Variety <= 0 {
		t.Errorf("expected variety > 0, got %d", breakdown.Variety)
	}
}

func TestCalculateScientificStreak(t *testing.T) {
	loc := time.UTC
	today := time.Date(2026, 8, 11, 0, 0, 0, 0, loc)

	// Case 1: Logged on time (CreatedAt is same day as Date)
	logs1 := []models.GymLog{
		{
			Date:       "2026-08-11",
			Hours:      1.5,
			CreatedAt:  time.Date(2026, 8, 11, 10, 0, 0, 0, loc),
			IsRestored: false,
		},
		{
			Date:       "2026-08-10",
			Hours:      1.5,
			CreatedAt:  time.Date(2026, 8, 10, 18, 0, 0, 0, loc),
			IsRestored: false,
		},
	}
	stats1 := CalculateScientificStreak(logs1, 3, 30, today)
	if stats1.CurrentStreak != 2 {
		t.Errorf("Case 1 (on-time): expected streak 2, got %d", stats1.CurrentStreak)
	}

	// Case 2: Past log created late (not restored)
	// Log for 2026-08-10 created on 2026-08-11
	logs2 := []models.GymLog{
		{
			Date:       "2026-08-11",
			Hours:      1.5,
			CreatedAt:  time.Date(2026, 8, 11, 10, 0, 0, 0, loc),
			IsRestored: false,
		},
		{
			Date:       "2026-08-10",
			Hours:      1.5,
			CreatedAt:  time.Date(2026, 8, 11, 12, 0, 0, 0, loc), // Late log! Created on Aug 11 for Aug 10
			IsRestored: false,
		},
	}
	stats2 := CalculateScientificStreak(logs2, 3, 30, today)
	// Since 2026-08-10 was logged late and is NOT restored, it shouldn't count towards streak.
	// So 2026-08-10 is treated as non-compliant, breaking the streak.
	// 2026-08-11 is compliant, so streak should be 1.
	if stats2.CurrentStreak != 1 {
		t.Errorf("Case 2 (late, not restored): expected streak 1, got %d", stats2.CurrentStreak)
	}

	// Case 3: Past log created late but RESTORED (IsRestored = true)
	logs3 := []models.GymLog{
		{
			Date:       "2026-08-11",
			Hours:      1.5,
			CreatedAt:  time.Date(2026, 8, 11, 10, 0, 0, 0, loc),
			IsRestored: false,
		},
		{
			Date:       "2026-08-10",
			Hours:      1.5,
			CreatedAt:  time.Date(2026, 8, 11, 12, 0, 0, 0, loc), // Late log but IsRestored is true
			IsRestored: true,
		},
	}
	stats3 := CalculateScientificStreak(logs3, 3, 30, today)
	if stats3.CurrentStreak != 2 {
		t.Errorf("Case 3 (late, restored): expected streak 2, got %d", stats3.CurrentStreak)
	}

	// Case 4: CreatedAt is Zero (seed/mock data)
	logs4 := []models.GymLog{
		{
			Date:       "2026-08-11",
			Hours:      1.5,
			IsRestored: false,
		},
		{
			Date:       "2026-08-10",
			Hours:      1.5,
			IsRestored: false,
		},
	}
	stats4 := CalculateScientificStreak(logs4, 3, 30, today)
	if stats4.CurrentStreak != 2 {
		t.Errorf("Case 4 (zero CreatedAt): expected streak 2, got %d", stats4.CurrentStreak)
	}
}
