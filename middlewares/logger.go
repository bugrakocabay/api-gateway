package middlewares

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"time"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestBody []byte
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				requestBody = bodyBytes
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		slog.Info("Incoming request", "method", r.Method, "path", r.URL.Path, "body", string(requestBody))

		rec := responseRecorder{ResponseWriter: w, body: new(bytes.Buffer), statusCode: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(&rec, r)
		duration := time.Since(start)

		// Log the response body and status code
		slog.Info("Outgoing response", "method", r.Method, "path", r.URL.Path, "status", rec.statusCode, "duration", duration, "body", rec.body.String())
	})
}

// responseRecorder is a custom ResponseWriter to capture the response body and status code
type responseRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (rec *responseRecorder) Write(b []byte) (int, error) {
	rec.body.Write(b)
	return rec.ResponseWriter.Write(b)
}

func (rec *responseRecorder) WriteHeader(statusCode int) {
	rec.statusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}
