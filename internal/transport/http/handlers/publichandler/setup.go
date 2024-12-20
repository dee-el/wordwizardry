package publichandler

import "net/http"

func SetupPublicRoutes(mux *http.ServeMux) {
	handler := NewPublicHandler()

	mux.HandleFunc("/connect", handler.IndexHandler)
	mux.HandleFunc("/leaderboard", handler.LeaderboardHandler)
}
