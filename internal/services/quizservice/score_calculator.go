package quizservice

const (
	baseScore     = 100 // base score for correct answer
	maxAnswerTime = 5.0 // maximum time in seconds
	minMultiplier = 0.1 // minimum score multiplier
	perfectTime   = 3.0 // time for maximum score
)

// Add scoring calculation function
func calculateScore(answerTime float64) int {
	if answerTime <= perfectTime {
		return baseScore
	}

	if answerTime >= maxAnswerTime {
		return int(float64(baseScore) * minMultiplier)
	}

	multiplier := 1.0 - ((answerTime-perfectTime)/(maxAnswerTime-perfectTime))*(1.0-minMultiplier)
	return int(float64(baseScore) * multiplier)
}
