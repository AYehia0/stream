package api

// the exposed server
import (
	"html/template"
	"os"
	"stream/internal/chat"
	"stream/pkg/logger"
)

type Server struct {
	logger     logger.Logger
	tmpl       *template.Template
	groqClient chat.GroqClient
}

func NewServer(logger logger.Logger, tmpl *template.Template) *Server {
	groqClient := chat.NewGroqClient(os.Getenv("GROQ_API_KEY"))
	return &Server{
		logger:     logger,
		tmpl:       tmpl,
		groqClient: groqClient,
	}
}
