package app

import (
	"stream/internal/api"
)

// reload the application routes
func (a *App) reloadRoutes() {

	appHandler := api.NewServer(a.logger, tmpl)

	a.router.HandleFunc("GET /status", appHandler.Status)
	a.router.HandleFunc("POST /chat", appHandler.SendMessage)
}
