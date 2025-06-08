package api

import (
	"net/http"
	"net/http/httptest"
	"stream/pkg/logger"
	"strings"
	"testing"
)

func TestLoggingMiddleware(t *testing.T) {
	// Temporarily redirect logger output to a buffer for test verification
	buf := &strings.Builder{}
	logger.SetOutput(buf)
	// Restore to default output after test
	defer logger.SetOutput(nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("I'm a teapot"))
	})

	wrapped := Logging(logger.Info, handler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusTeapot {
		t.Fatalf("expected status %d, got %d", http.StatusTeapot, res.StatusCode)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "/test") || !strings.Contains(logOutput, "418") {
		t.Fatalf("log output missing expected content: %s", logOutput)
	}
}
