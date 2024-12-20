package quizservice

import "wordwizardry/internal/pkg/models"

type JoinQuizRequest struct {
	QuizID   string `json:"quiz_id"`
	Username string `json:"username"`
}

type JoinQuizResponse struct {
	SessionID string            `json:"session_id"`
	PlayerID  string            `json:"player_id"`
	Questions []models.Question `json:"questions"`
}
