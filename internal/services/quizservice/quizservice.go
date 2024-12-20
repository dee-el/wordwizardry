package quizservice

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"wordwizardry/internal/pkg/models"

	"wordwizardry/internal/services/broadcast"
	"wordwizardry/internal/services/quizservice/sessions"

	"wordwizardry/internal/services/quizservice/quizrepositories"
)

type QuizService struct {
	quizReader quizrepositories.QuizReader
	quizWriter quizrepositories.QuizWriter

	sessionManager sessions.SessionManager
	hub            broadcast.Hub
}

func NewQuizService(
	quizReader quizrepositories.QuizReader,
	quizWriter quizrepositories.QuizWriter,

	sessionManager sessions.SessionManager,
	hub broadcast.Hub,
) *QuizService {
	return &QuizService{
		quizReader: quizReader,
		quizWriter: quizWriter,

		sessionManager: sessionManager,
		hub:            hub,
	}
}

func (s *QuizService) JoinQuiz(ctx context.Context, req JoinQuizRequest) (*JoinQuizResponse, error) {
	quiz, questions, err := s.quizReader.GetQuiz(ctx, req.QuizID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quiz: %w", err)
	}

	if quiz == nil || quiz.Status != models.QuizStatusActive || len(questions) == 0 {
		return nil, fmt.Errorf("quiz not found")
	}

	session, err := s.sessionManager.FindQuizSessionByQuizID(ctx, req.QuizID)
	if err != nil {
		return nil, fmt.Errorf("failed to find session: %w", err)
	}

	if session == nil {
		session = &models.Session{
			Quiz:      quiz,
			Questions: questions,
			Players:   []models.SessionPlayer{},
			Result:    make(map[string]models.Answer),
		}

		err = s.sessionManager.CreateQuizSession(ctx, session)
		if err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}

		if err := s.hub.CreateRoom(session.ID); err != nil {
			return nil, fmt.Errorf("failed to create room: %w", err)
		}
	}

	player := models.Player{
		ID:       uuid.New().String(),
		Username: req.Username,
	}

	sessionPlayer := models.SessionPlayer{
		Player: player,
		QuizID: req.Username,
		Score:  0,
	}

	err = s.sessionManager.AddPlayerToQuizSession(ctx, session.ID, sessionPlayer)
	if err != nil {
		return nil, fmt.Errorf("failed to add player to session: %w", err)
	}

	err = s.hub.JoinRoom(session.ID, player.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to join room: %w", err)
	}

	return &JoinQuizResponse{
		SessionID: session.ID,
		PlayerID:  player.ID,
		Questions: questions,
	}, nil
}

type SubmitAnswerRequest struct {
	PlayerID            string  `json:"player_id"`
	SessionID           string  `json:"session_id"`
	QuizID              string  `json:"quiz_id"`
	QuestionID          string  `json:"question_id"`
	Answer              string  `json:"answer"`
	AnswerTimeInSeconds float64 `json:"answer_time"`
}

func (s *QuizService) SubmitAnswer(ctx context.Context, req SubmitAnswerRequest) error {
	session, err := s.sessionManager.FindQuizPlayerSession(ctx, req.SessionID, req.PlayerID)
	if err != nil {
		return fmt.Errorf("failed to find session: %w", err)
	}

	if session == nil {
		return fmt.Errorf("session not found")
	}

	question := models.Question{}
	for _, q := range session.Questions {
		if q.ID == req.QuestionID {
			question = q
			break
		}
	}

	if question.ID == "" {
		return fmt.Errorf("question not found")
	}

	resKey := fmt.Sprintf("%s:%s", req.QuestionID, req.PlayerID)

	hasAnswered := false
	_, ok := session.Result[resKey]
	if ok {
		hasAnswered = true
	}

	if hasAnswered {
		return fmt.Errorf("question already answered")
	}

	correct := false
	for _, option := range question.Options {
		if option == req.Answer {
			correct = true
			break
		}
	}

	var score int
	if correct {
		score = calculateScore(req.AnswerTimeInSeconds)
	}

	err = s.sessionManager.UpdateQuizPlayerScoreSession(
		ctx,
		req.SessionID,
		req.PlayerID,
		score,
		models.Result{
			resKey: models.Answer{
				PlayerChoice:  req.Answer,
				CorrectAnswer: question.Correct,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to add score: %w", err)
	}

	leaderboard, err := s.sessionManager.FindLeaderboardQuizSession(ctx, session.ID)
	if err != nil {
		return fmt.Errorf("failed to get leaderboard: %w", err)
	}

	// Broadcast to room instead of session
	messages := []models.WSMessage{
		{
			Type: "answer_submitted",
			Data: map[string]interface{}{
				"player_id": req.PlayerID,
				"correct":   correct,
				"score":     score,
			},
		},
		{
			Type: "leaderboard_update",
			Data: map[string]interface{}{
				"leaderboard": leaderboard,
			},
		},
	}

	// Broadcast all messages to the room
	for _, msg := range messages {
		if err := s.hub.BroadcastToRoom(ctx, session.ID, msg); err != nil {
			return fmt.Errorf("failed to broadcast message: %w", err)
		}
	}

	return nil
}

func (s *QuizService) ValidatePlayerSession(ctx context.Context, sessionID, playerID string) (*models.Session, error) {
	session, err := s.sessionManager.FindQuizPlayerSession(ctx, sessionID, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find session: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	return session, nil
}
