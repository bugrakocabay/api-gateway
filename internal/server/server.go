package server

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/bugrakocabay/api-gateway/internal/config"
	"github.com/bugrakocabay/api-gateway/internal/middlewares"
	"github.com/bugrakocabay/api-gateway/internal/routes"
	"github.com/gorilla/mux"
)

const PORT = 8080

// Start initializes and starts the HTTP server.
func Start(cfg *config.Config) *http.Server {
	router := mux.NewRouter()
	localThrottler := middlewares.NewLocalThrottler(cfg.Routes)
	router.Use(localThrottler.LocalThrottlerMiddleware)
	router.Use(middlewares.InjectRequestID)
	router.Use(middlewares.Logger)

	for _, route := range cfg.Routes {
		handler := routes.CreateHandler(*route)
		router.HandleFunc(route.Path, handler).Methods(route.Method)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", PORT),
		Handler: router,
	}

	go func() {
		slog.Info("Starting API Gateway", "port", PORT)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %s", err)
		}
	}()

	return srv
}
