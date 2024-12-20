package quizhandler

import (
	"net/http"

	"wordwizardry/internal/services/broadcast"
	"wordwizardry/internal/services/quizservice"
)

func SetupQuizRoutes(
	mux *http.ServeMux,
	quizService *quizservice.QuizService,
	hub *broadcast.WebSocketHub,
) {
	handler := NewQuizHandler(quizService, hub)

	mux.HandleFunc("/api/quiz/join", handler.JoinQuiz)
	mux.HandleFunc("/api/quiz/submit-answer", handler.SubmitAnswer)
	mux.HandleFunc("/ws", handler.HandleWebSocket)
}
