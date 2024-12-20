package quizhandler

import (
	"encoding/json"
	"net/http"

	"wordwizardry/internal/services/broadcast"
	"wordwizardry/internal/services/quizservice"
)

type QuizHandler struct {
	quizService *quizservice.QuizService
	hub         broadcast.Hub
}

func NewQuizHandler(quizService *quizservice.QuizService, hub broadcast.Hub) *QuizHandler {
	return &QuizHandler{
		quizService: quizService,
		hub:         hub,
	}
}

func (h *QuizHandler) JoinQuiz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req quizservice.JoinQuizRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.QuizID == "" || req.Username == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	resp, err := h.quizService.JoinQuiz(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *QuizHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req quizservice.SubmitAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.PlayerID == "" || req.SessionID == "" || req.QuizID == "" || req.QuestionID == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	if err := h.quizService.SubmitAnswer(r.Context(), req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
