package service

import (
	"strings"

	"gymgit/backend/internal/models"
)

// CalculateSplitAccuracy evaluates workout split accuracy score (0-100%) for a 7-day cycle.
// It compares prescribed plan categories against actual user logged workout types within the 7-day window.
func CalculateSplitAccuracy(logs []models.GymLog, plan *models.WeeklyPlan, cycleStartDate string, cycleEndDate string) models.AccuracyBreakdown {
	if plan == nil || len(plan.Categories) == 0 {
		return models.AccuracyBreakdown{
			AccuracyScore:  100,
			TotalEvaluated: 0,
			MatchedCount:   0,
		}
	}

	// Filter logs within the 7-day cycle window
	cycleLogsMap := make(map[string]string) // date -> workoutType
	for _, l := range logs {
		if l.Date >= cycleStartDate && l.Date <= cycleEndDate && l.Hours > 0 {
			cycleLogsMap[l.Date] = l.WorkoutType
		}
	}

	if len(cycleLogsMap) == 0 {
		return models.AccuracyBreakdown{
			AccuracyScore:  100,
			TotalEvaluated: 0,
			MatchedCount:   0,
		}
	}

	matchedCount := 0
	totalEvaluated := len(cycleLogsMap)
	var deviations []string

	// For each active logged date in the cycle, check if the workout type matches any prescribed category
	for dateStr, actualType := range cycleLogsMap {
		matched := false
		for _, prescribed := range plan.Categories {
			if strings.EqualFold(actualType, prescribed) {
				matched = true
				break
			}
		}
		if matched {
			matchedCount++
		} else {
			deviations = append(deviations, dateStr+": logged "+actualType+" not in prescribed split")
		}
	}

	accuracyScore := 100
	if totalEvaluated > 0 {
		ratio := float64(matchedCount) / float64(totalEvaluated)
		accuracyScore = int(ratio * 100.0)
	}

	return models.AccuracyBreakdown{
		AccuracyScore:  accuracyScore,
		TotalEvaluated: totalEvaluated,
		MatchedCount:   matchedCount,
		Deviations:     deviations,
	}
}
