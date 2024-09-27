package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bugrakocabay/api-gateway/internal/config"
	"github.com/bugrakocabay/api-gateway/internal/server"
)

const ConfigPath = "API_GATEWAY_CONFIG_PATH"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	cfg, err := config.LoadConfig(os.Getenv(ConfigPath))
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	srv := server.Start(cfg)

	gracefulShutdown(srv, 15*time.Second)
}

func gracefulShutdown(srv *http.Server, timeout time.Duration) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	slog.Info("Shutdown complete")
}
