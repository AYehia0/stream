package app

import (
	"stream/internal/api"
)

func (a *App) reloadRoutes() {

	appHandler := api.NewServer(a.logger)

	a.router.HandleFunc("GET /status", appHandler.Status)
	a.router.HandleFunc("POST /chat", appHandler.SendMessage)
}
