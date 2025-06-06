package api

import (
	"net/http"
	"stream/pkg/logger"
	"time"
)

// Middleware represents the type signature of a middleware
// function.
type Middleware func(http.Handler) http.Handler

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func Logging(logger logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &wrappedWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		logger.Printf("%s %s %d %s %s",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			time.Since(start),
			r.RemoteAddr,
		)
	})
}
