package service_test

import (
	"testing"

	"gymgit/backend/internal/models"
	"gymgit/backend/internal/service"
)

func TestCalculateSplitAccuracy(t *testing.T) {
	plan := &models.WeeklyPlan{
		ID:         "ppl-standard",
		Name:       "Push Pull Legs",
		Categories: []string{"Push", "Pull", "Legs", "Cardio"},
	}

	logs := []models.GymLog{
		{Date: "2026-08-01", Hours: 1.0, WorkoutType: "Push"},
		{Date: "2026-08-02", Hours: 1.2, WorkoutType: "Pull"},
		{Date: "2026-08-03", Hours: 1.5, WorkoutType: "Legs"},
		{Date: "2026-08-04", Hours: 0.8, WorkoutType: "Cardio"},
	}

	breakdown := service.CalculateSplitAccuracy(logs, plan, "2026-08-01", "2026-08-07")
	if breakdown.AccuracyScore != 100 {
		t.Errorf("expected 100%% accuracy, got %d", breakdown.AccuracyScore)
	}

	// Add unprescribed workout type
	logs = append(logs, models.GymLog{Date: "2026-08-05", Hours: 1.0, WorkoutType: "UnrelatedActivity"})
	breakdownPartial := service.CalculateSplitAccuracy(logs, plan, "2026-08-01", "2026-08-07")
	if breakdownPartial.AccuracyScore >= 100 {
		t.Errorf("expected < 100%% accuracy for unprescribed activity, got %d", breakdownPartial.AccuracyScore)
	}
	if len(breakdownPartial.Deviations) != 1 {
		t.Errorf("expected 1 deviation, got %d", len(breakdownPartial.Deviations))
	}
}
