package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

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
		return nil, err
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
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

	log.Printf("Starting API Gateway on PORT: %d", PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", PORT), router))
}
