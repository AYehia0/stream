package app

import (
	"stream/internal/api"

	httpSwagger "github.com/swaggo/http-swagger"
)

func (a *App) reloadRoutes(appHandler *api.Handler) {
	a.router.HandleFunc("GET /swagger/*", httpSwagger.WrapHandler)
	a.router.HandleFunc("GET /status", appHandler.Status)
	a.router.HandleFunc("POST /chat", appHandler.SendMessage)
}
