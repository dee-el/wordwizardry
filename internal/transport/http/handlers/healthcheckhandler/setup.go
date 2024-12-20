package healthcheckhandler

import "net/http"

func SetupHealthCheckRoutes(mux *http.ServeMux) {
	handler := NewHealthHandler()

	mux.HandleFunc("/health", handler.ServeHTTP)
}
