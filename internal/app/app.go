package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"stream/internal/api"
	"stream/pkg/logger"
	"time"
)

type App struct {
	logger logger.Logger
	router *http.ServeMux
}

func New(logger logger.Logger) *App {
	return &App{
		logger: logger,
		router: http.NewServeMux(),
	}
}

func (a *App) Run(ctx context.Context) error {

	a.reloadRoutes()

	server := &http.Server{
		Addr:    ":8080",
		Handler: api.Logging(a.logger, a.router),
	}

	done := make(chan struct{})
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Fatalf("server error", slog.Any("Error", err))
		}
		close(done)
	}()

	a.logger.Printf("server started", slog.String("address", server.Addr))

	select {
	case <-done:
		break

	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		server.Shutdown(ctx)
		cancel()
	}

	return nil

}
