package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

const PORT = 8080

type Route struct {
	Path   string `json:"path"`
	Target string `json:"target"`
	Method string `json:"method"`
}

type Config struct {
	Routes []Route `json:"routes"`
}

func loadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	for i, route := range config.Routes {
		if route.Path == "" {
			return nil, fmt.Errorf("route %d has empty path", i)
		}
		if route.Target == "" {
			return nil, fmt.Errorf("route %d has empty target", i)
		}
		if route.Method == "" {
			return nil, fmt.Errorf("route %d has empty method", i)
		}
	}

	return &config, nil
}

func createHandler(route Route) http.HandlerFunc {
	target, err := url.Parse(route.Target)
	if err != nil {
		log.Printf("Invalid Target URL for route %s: %v", route.Path, err)
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Invalid Target URL", http.StatusInternalServerError)
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	return func(w http.ResponseWriter, r *http.Request) {
		done := make(chan bool)
		go func() {
			defer close(done)
			log.Printf("Proxying request: %s %s -> %s", r.Method, r.URL.Path, target)
			proxy.ServeHTTP(w, r)
		}()
		<-done
	}
}

func main() {
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	router := mux.NewRouter()

	for _, route := range config.Routes {
		router.HandleFunc(route.Path, createHandler(route)).Methods(route.Method)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", PORT),
		Handler: router,
	}

	go func() {
		log.Printf("Starting API Gateway on PORT: %d", PORT)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
