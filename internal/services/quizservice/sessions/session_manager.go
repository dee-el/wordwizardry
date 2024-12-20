package sessions

import (
	"context"
	"wordwizardry/internal/pkg/models"
)

type SessionManager interface {
	FindQuizSession(ctx context.Context, sessionID string) (*models.Session, error)
	FindQuizSessionByQuizID(ctx context.Context, quizID string) (*models.Session, error)
	CreateQuizSession(ctx context.Context, session *models.Session) error

	AddPlayerToQuizSession(ctx context.Context, sessionID string, player models.SessionPlayer) error

	FindQuizPlayerSession(ctx context.Context, sessionID, playerID string) (*models.Session, error)

	// should update the leaderboard
	UpdateQuizPlayerScoreSession(ctx context.Context, sessionID, playerID string, score int, res models.Result) error

	// sorted by score
	FindLeaderboardQuizSession(ctx context.Context, quizSessionID string) ([]models.SessionPlayer, error)
}
