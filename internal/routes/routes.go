package routes

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/bugrakocabay/api-gateway/internal/config"
	"github.com/gorilla/mux"
)

// CreateHandler returns an HTTP handler for a given route.
func CreateHandler(route config.Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		targetURLStr := route.Target
		for key, value := range vars {
			placeholder := "{" + key + "}"
			targetURLStr = strings.ReplaceAll(targetURLStr, placeholder, value)
		}

		targetURL, err := url.Parse(targetURLStr)
		if err != nil {
			slog.Error("Invalid Target URL after variable substitution", "error", err)
			http.Error(w, "Invalid Target URL", http.StatusInternalServerError)
			return
		}

		proxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = targetURL.Scheme
				req.URL.Host = targetURL.Host
				req.URL.Path = targetURL.Path
				req.Host = targetURL.Host

				req.URL.RawQuery = r.URL.RawQuery

				req.Header = r.Header
			},
			ErrorHandler: func(w http.ResponseWriter, req *http.Request, err error) {
				http.Error(w, "Proxy error: "+err.Error(), http.StatusBadGateway)
			},
		}

		proxy.ServeHTTP(w, r)
	}
}
