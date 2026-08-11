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
