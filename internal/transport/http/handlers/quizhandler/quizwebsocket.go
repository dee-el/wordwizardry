package quizhandler

import (
	"log"
	"net/http"

	"wordwizardry/internal/pkg/models"
)

func (h *QuizHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	playerID := r.URL.Query().Get("player_id")

	if sessionID == "" || playerID == "" {
		http.Error(w, "Missing session_id or player_id", http.StatusBadRequest)
		return
	}

	// Validate session and player
	session, err := h.quizService.ValidatePlayerSession(r.Context(), sessionID, playerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get player username
	uname := ""
	for _, p := range session.Players {
		if p.ID == playerID {
			uname = p.Username
		}
	}

	// Handle WebSocket connection
	err = h.hub.HandleWebSocket(w, r)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}

	// Send welcome message to the player
	welcomeMsg := models.WSMessage{
		Type: "room_joined",
		Data: map[string]interface{}{
			"room_info": map[string]interface{}{
				"player_count": len(session.Players),
				"leaderboard":  session.Players,
			},
			"player": map[string]interface{}{
				"id":       playerID,
				"username": uname,
			},
		},
	}

	// Broadcast player joined to room
	joinMsg := models.WSMessage{
		Type: "player_connected",
		Data: map[string]interface{}{
			"count": len(session.Players),
			"player": map[string]interface{}{
				"id":       playerID,
				"username": uname,
			},
		},
	}

	// Send messages
	err = h.hub.SendToPlayer(r.Context(), sessionID, playerID, welcomeMsg)
	if err != nil {
		log.Printf("Failed to send welcome message: %v", err)
	}

	err = h.hub.BroadcastToRoom(r.Context(), sessionID, joinMsg)
	if err != nil {
		log.Printf("Failed to broadcast player joined: %v", err)
	}
}
