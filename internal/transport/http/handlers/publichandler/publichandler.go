package publichandler

import (
	"fmt"
	"net/http"
	"text/template"
)

type PublicHandler struct{}

func NewPublicHandler() *PublicHandler {
	return &PublicHandler{}
}

func (p *PublicHandler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("public/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (p *PublicHandler) LeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	sessionID := r.URL.Query().Get("session_id")

	if playerID == "" || sessionID == "" {
		http.Error(w, "Both player_id and session_id are required", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Connected with player_id: %s and session_id: %s", playerID, sessionID)
}
