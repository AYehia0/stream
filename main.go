package main

import (
	"context"
	"os"
	"os/signal"
	"stream/internal/app"
	"stream/internal/config"
	"stream/internal/persistence"
	"stream/pkg/logger"
)

//	@title			Stream Service API
//	@version		1.0
//	@description	This is a sample server for a Groq stream service.

// @host	localhost:8080
func main() {
	err := config.ReadEnv()
	if err != nil {
		logger.Error.Printf("failed to read environment variables: %v", err)
		os.Exit(1)
	}

	log := logger.NewStdLogger(logger.Info)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	db := persistence.NewPersistence(persistence.MemoryStorage)

	a := app.New(log, db)

	if err := a.Run(ctx); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
