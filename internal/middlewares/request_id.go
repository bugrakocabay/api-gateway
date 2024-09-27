package middlewares

import (
	"net/http"

	"github.com/google/uuid"
)

const RequestIdHeader = "X-Request-ID"

// InjectRequestID appends a unique UUID to request header with a key named 'X-Request-ID'.
func InjectRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Add(RequestIdHeader, uuid.New().String())
		next.ServeHTTP(w, r)
	})
}
