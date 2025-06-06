package app

import (
	"html/template"
	"stream/internal/api"
)

var TemplatePath = "./templates/*"

// reload the application routes
func (a *App) reloadRoutes() {
	tmpl := template.Must(template.New("").ParseGlob("./templates/*"))

	appHandler := api.NewServer(a.logger, tmpl)

	// handle the static files
	// files := http.FileServer(http.Dir("./static"))
	// a.router.Handle("GET /static/", http.StripPrefix("/static/", files))
	a.router.HandleFunc("GET /{$}", appHandler.Index)
	a.router.HandleFunc("GET /status", appHandler.Status)
	a.router.HandleFunc("GET /chat", appHandler.SendMessage)
}
