package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wordwizardry/internal/transport/http/handlers/healthcheckhandler"
	"wordwizardry/internal/transport/http/handlers/publichandler"
	"wordwizardry/internal/transport/http/handlers/quizhandler"

	"wordwizardry/internal/services/broadcast"

	"wordwizardry/internal/services/quizservice"
	inmemory "wordwizardry/internal/services/quizservice/quizrepositories/inmemory"
	redissessionmanager "wordwizardry/internal/services/quizservice/sessions/redis"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	sessionManager, err := redissessionmanager.NewRedisSessionManager(redisURL)
	if err != nil {
		return err
	}

	hub := broadcast.NewWebSocketHub()
	go hub.Run()

	quizReader := inmemory.NewQuizRepository()
	quizWriter := inmemory.NewQuizRepository()

	quizService := quizservice.NewQuizService(
		quizReader,
		quizWriter,
		sessionManager,
		hub,
	)

	mux := http.NewServeMux()

	healthcheckhandler.SetupHealthCheckRoutes(mux)
	publichandler.SetupPublicRoutes(mux)
	quizhandler.SetupQuizRoutes(mux, quizService, hub)

	// Create server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
		// Good practice settings
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("Server listening on %s", srv.Addr)
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking select waiting for either a server error or a system signal
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Printf("Starting shutdown... Signal: %v", sig)

		// Give outstanding requests 15 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Shutdown the server
		if err := srv.Shutdown(ctx); err != nil {
			// If shutdown times out, force close
			srv.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
