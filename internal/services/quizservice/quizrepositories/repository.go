package quizrepositories

import (
	"context"

	"wordwizardry/internal/pkg/models"
)

type QuizWriter interface {
	CreateQuiz(ctx context.Context, quiz *models.Quiz) error
	UpdateQuiz(ctx context.Context, quiz *models.Quiz) error
	MapQuestions(ctx context.Context, quizID string, questions []models.Question) error
	SaveQuizResult(ctx context.Context, quizResult *models.QuizResult) error
}

type QuizReader interface {
	GetQuiz(ctx context.Context, id string) (*models.Quiz, []models.Question, error)
}
