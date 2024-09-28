package middlewares

import (
	"net/http"

	"github.com/google/uuid"
)

const TraceIdHeader = "X-Trace-ID"

// InjectTraceID appends a unique UUID to request header with a key named 'X-Request-ID'.
func InjectTraceID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Add(TraceIdHeader, uuid.New().String())
		next.ServeHTTP(w, r)
	})
}
