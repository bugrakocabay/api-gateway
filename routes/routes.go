package routes

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/bugrakocabay/api-gateway/config"
)

// CreateHandler returns an HTTP handler for a given route.
func CreateHandler(route config.Route) http.HandlerFunc {
	targetURL, err := url.Parse(route.Target)
	if err != nil {
		slog.Error("Invalid Target URL", "route", route.Path, "error", err)
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Invalid Target URL", http.StatusInternalServerError)
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}
