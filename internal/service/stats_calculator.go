package service

import (
	"math"
	"strings"
	"time"

	"gymgit/backend/internal/models"
)

// CalculateTargetDaysPerWeek returns targetDaysPerWeek = min(6, max(3, activePlanCategoriesCount))
func CalculateTargetDaysPerWeek(categories []string) int {
	activeCategoriesCount := 0
	for _, cat := range categories {
		if !strings.EqualFold(cat, "Rest") && cat != "" {
			activeCategoriesCount++
		}
	}
	if activeCategoriesCount == 0 {
		activeCategoriesCount = len(categories)
	}

	target := activeCategoriesCount
	if target < 3 {
		target = 3
	}
	if target > 6 {
		target = 6
	}
	return target
}

// IsDateCompliant checks if date D has logged hours > 0 OR window [D-6, D] active sessions >= max(2, targetDaysPerWeek - 1)
func IsDateCompliant(targetDate time.Time, logsMap map[string]float64, targetDaysPerWeek int) bool {
	dateStr := targetDate.Format("2006-01-02")
	if hours, exists := logsMap[dateStr]; exists && hours > 0 {
		return true
	}

	// Calculate active sessions in rolling 7-day window [D-6, D]
	activeInWindow := 0
	for i := 0; i < 7; i++ {
		checkDate := targetDate.AddDate(0, 0, -i).Format("2006-01-02")
		if hours, exists := logsMap[checkDate]; exists && hours > 0 {
			activeInWindow++
		}
	}

	requiredSessions := targetDaysPerWeek - 1
	if requiredSessions < 2 {
		requiredSessions = 2
	}

	return activeInWindow >= requiredSessions
}

// CalculateScientificStreak calculates current streak and compliance rate
func CalculateScientificStreak(logs []models.GymLog, targetDaysPerWeek int, daysWindow int) models.StreakStats {
	if daysWindow <= 0 {
		daysWindow = 30
	}

	logsMap := make(map[string]float64)
	for _, l := range logs {
		if l.Hours > 0 {
			logsMap[l.Date] = l.Hours
		}
	}

	today := time.Now().UTC()

	// 1. Current Streak: count backward from today until non-compliant day
	currentStreak := 0
	currDate := today
	for {
		if IsDateCompliant(currDate, logsMap, targetDaysPerWeek) {
			currentStreak++
			currDate = currDate.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	// 2. Compliance Rate over daysWindow
	totalCompliantDays := 0
	totalActiveDays := 0
	for i := 0; i < daysWindow; i++ {
		d := today.AddDate(0, 0, -i)
		dStr := d.Format("2006-01-02")
		if IsDateCompliant(d, logsMap, targetDaysPerWeek) {
			totalCompliantDays++
		}
		if hours, exists := logsMap[dStr]; exists && hours > 0 {
			totalActiveDays++
		}
	}

	complianceRate := int(math.Round((float64(totalCompliantDays) / float64(daysWindow)) * 100.0))

	return models.StreakStats{
		CurrentStreak:      currentStreak,
		TargetDaysPerWeek:  targetDaysPerWeek,
		ComplianceRate:     complianceRate,
		TotalCompliantDays: totalCompliantDays,
		TotalTrackedDays:   daysWindow,
		TotalActiveDays:    totalActiveDays,
	}
}

// CalculatePowerScore calculates the 4 Gym Power Score components (0-100 total)
func CalculatePowerScore(logs []models.GymLog, targetDaysPerWeek int, periodTotalDays int) models.PowerScoreBreakdown {
	if periodTotalDays <= 0 {
		periodTotalDays = 30
	}

	today := time.Now().UTC()
	startDate := today.AddDate(0, 0, -periodTotalDays+1).Format("2006-01-02")

	// Filter logs within window and calculate session qualities
	activeDaysMap := make(map[string]float64)
	uniqueWorkoutTypesMap := make(map[string]bool)
	var sessionQualities []float64

	for _, l := range logs {
		if l.Date >= startDate && l.Date <= today.Format("2006-01-02") && l.Hours > 0 {
			activeDaysMap[l.Date] = l.Hours
			if l.WorkoutType != "" && !strings.EqualFold(l.WorkoutType, "Rest") {
				uniqueWorkoutTypesMap[strings.ToLower(l.WorkoutType)] = true
			}

			// Session Duration Quality
			q := 1.0
			h := l.Hours
			if h > 1.75 {
				q = math.Max(0.4, 1.0-(h-1.75)*0.25)
			} else if h < 0.75 {
				q = math.Max(0.2, h/0.75)
			}
			sessionQualities = append(sessionQualities, q)
		}
	}

	activeDaysCount := len(activeDaysMap)
	uniqueWorkoutTypesCount := len(uniqueWorkoutTypesMap)

	// A. Consistency (0-45 pts)
	targetActiveDays := math.Round((float64(periodTotalDays) / 7.0) * float64(targetDaysPerWeek))
	if targetActiveDays < 1.0 {
		targetActiveDays = 1.0
	}
	consistencyRatio := math.Min(1.0, float64(activeDaysCount)/targetActiveDays)
	consistencyScore := int(math.Round(consistencyRatio * 45.0))

	// B. Duration Quality (0-25 pts)
	durationScore := 0
	if len(sessionQualities) > 0 {
		sumQuality := 0.0
		for _, q := range sessionQualities {
			sumQuality += q
		}
		avgQuality := sumQuality / float64(len(sessionQualities))
		durationScore = int(math.Round(avgQuality * 25.0))
	}

	// C. Variety (0-20 pts)
	varietyRatio := math.Min(1.0, float64(uniqueWorkoutTypesCount)/3.0)
	varietyScore := int(math.Round(varietyRatio * 20.0))

	// D. Momentum (0-10 pts)
	momentumTarget := float64(periodTotalDays) * 0.5
	if momentumTarget < 1.0 {
		momentumTarget = 1.0
	}
	momentumRatio := math.Min(1.0, float64(activeDaysCount)/momentumTarget)
	momentumScore := int(math.Round(momentumRatio * 10.0))

	totalScore := consistencyScore + durationScore + varietyScore + momentumScore
	if totalScore < 0 {
		totalScore = 0
	}
	if totalScore > 100 {
		totalScore = 100
	}

	return models.PowerScoreBreakdown{
		Consistency:     consistencyScore,
		DurationQuality: durationScore,
		Variety:         varietyScore,
		Momentum:        momentumScore,
		TotalScore:      totalScore,
	}
}

// MapPowerScoreToAnimeTier maps a Gym Power Score (0-100) to its corresponding Anime Tier
func MapPowerScoreToAnimeTier(score int) models.AnimeTier {
	switch {
	case score >= 95:
		return models.AnimeTier{
			Tier:      "SS",
			Character: "Saitama",
			Anime:     "One Punch Man",
			Title:     "One Punch God",
		}
	case score >= 85:
		return models.AnimeTier{
			Tier:      "S+",
			Character: "Satoru Gojo",
			Anime:     "Jujutsu Kaisen",
			Title:     "The Honored One",
		}
	case score >= 70:
		return models.AnimeTier{
			Tier:      "S",
			Character: "Monkey D. Luffy",
			Anime:     "One Piece",
			Title:     "Gear 5 Sun God Nika",
		}
	case score >= 55:
		return models.AnimeTier{
			Tier:      "A",
			Character: "Izuku Midoriya",
			Anime:     "My Hero Academia",
			Title:     "One For All Successor",
		}
	case score >= 35:
		return models.AnimeTier{
			Tier:      "B",
			Character: "Tanjiro Kamado",
			Anime:     "Demon Slayer",
			Title:     "Water Breathing Swordsman",
		}
	case score >= 15:
		return models.AnimeTier{
			Tier:      "C",
			Character: "Mumen Rider",
			Anime:     "One Punch Man",
			Title:     "Class-C Hero of Justice",
		}
	default:
		return models.AnimeTier{
			Tier:      "D",
			Character: "Aqua",
			Anime:     "Konosuba",
			Title:     "Useless Goddess",
		}
	}
}
