package api

// the exposed server
import (
	"os"
	"stream/internal/chat"
	"stream/pkg/logger"
)

type Server struct {
	logger     logger.Logger
	groqClient chat.GroqClient
}

func NewServer(logger logger.Logger) *Server {
	groqClient := chat.NewGroqClient(os.Getenv("GROQ_API_KEY"))
	return &Server{
		logger:     logger,
		groqClient: groqClient,
	}
}
