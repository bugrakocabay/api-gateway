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
		targetURL, err := parseTargetURL(vars, route.Target)
		if err != nil {
			slog.Error("Error occurred while parsing target URL ", "error", err)
			http.Error(w, "Invalid Target URL", http.StatusInternalServerError)
			return
		}

		proxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = targetURL.Scheme
				req.URL.Host = targetURL.Host
				req.URL.Path = targetURL.Path
				req.Host = targetURL.Host
				req.Method = route.Target.Method

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

func parseTargetURL(vars map[string]string, target *config.Target) (*url.URL, error) {
	targetPath := target.Path
	targetHost := target.Host
	for key, value := range vars {
		placeholder := "{" + key + "}"
		targetPath = strings.ReplaceAll(targetPath, placeholder, value)
	}

	targetURL, err := url.Parse(targetHost + targetPath)
	if err != nil {
		return nil, err
	}

	return targetURL, nil
}
