package routes

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/bugrakocabay/bifrost/internal/config"
	"github.com/gorilla/mux"
)

// CreateHandler returns an HTTP handler for a given route.
// This function sets up the reverse proxy logic for Bifrost. It precomputes
// the allowed query parameters for the route and creates an HTTP handler function
// that processes incoming requests, modifies them according to the route configuration,
// and forwards them to the appropriate target URL using a reverse proxy.
func CreateHandler(route config.Route) http.HandlerFunc {
	allowedParamsSet := make(map[string]struct{}, len(route.QueryParams))
	allowAll := false
	for _, qp := range route.QueryParams {
		if qp == "*" {
			allowAll = true
			break
		}
		allowedParamsSet[qp] = struct{}{}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		targetURL, err := parseTargetURL(r, allowedParamsSet, allowAll, route.Target)
		if err != nil {
			slog.Error("Error parsing target URL", "route", route.Path, "error", err)
			http.Error(w, "Invalid Target URL", http.StatusInternalServerError)
			return
		}

		proxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = targetURL.Scheme
				req.URL.Host = targetURL.Host
				req.URL.Path = targetURL.Path
				req.URL.RawQuery = targetURL.RawQuery
				req.Host = targetURL.Host

				if route.Target.Method != "" {
					req.Method = route.Target.Method
				}

				copyHeaders(req.Header, r.Header)
			},
			ErrorHandler: func(w http.ResponseWriter, req *http.Request, err error) {
				http.Error(w, "Proxy error: "+err.Error(), http.StatusBadGateway)
			},
		}

		proxy.ServeHTTP(w, r)
	}
}

// parseTargetURL constructs the final target URL for the outgoing proxy request.
// It replaces any path variables (like {userId}) with actual values from the request,
// filters the query parameters based on the allowed list, and constructs the full URL
// that the proxy will use to forward the request.
func parseTargetURL(r *http.Request, allowedParamsSet map[string]struct{}, allowAll bool, target *config.Target) (*url.URL, error) {
	vars := mux.Vars(r)
	targetPath := replacePathVariables(target.Path, vars)

	targetBaseURL, err := url.Parse(target.Host)
	if err != nil {
		return nil, fmt.Errorf("invalid target host URL: %w", err)
	}

	targetURL := targetBaseURL.ResolveReference(&url.URL{Path: targetPath})

	if allowAll {
		targetURL.RawQuery = r.URL.RawQuery
	} else {
		originalQuery := r.URL.Query()
		filteredQuery := url.Values{}
		for key, values := range originalQuery {
			if _, allowed := allowedParamsSet[key]; allowed {
				filteredQuery[key] = values
			}
		}
		targetURL.RawQuery = filteredQuery.Encode()
	}

	return targetURL, nil
}

func replacePathVariables(path string, vars map[string]string) string {
	for key, value := range vars {
		placeholder := "{" + key + "}"
		path = strings.ReplaceAll(path, placeholder, value)
	}
	return path
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
