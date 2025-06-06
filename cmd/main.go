package main

import (
	"context"
	"os"
	"os/signal"
	"stream/internal/app"
	"stream/internal/config"
	"stream/pkg/logger"
)

func main() {
	config.ReadEnv()

	log := logger.NewStdLogger(logger.Info)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	a := app.New(log)

	if err := a.Run(ctx); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
