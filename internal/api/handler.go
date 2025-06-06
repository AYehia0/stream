package api

import (
	"net/http"
	"os"
	"strconv"
	"stream/internal/chat"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages    []ChatMessage `json:"messages"`
	Model       string        `json:"model"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_completion_tokens"`
	/*
		Use top_p < 1.0 (e.g. 0.85) if you want less rambling and more concise completions.

		Use top_p = 1.0 if you’re relying on maximum diversity.

		Keep temperature and top_p balanced — don’t set both to extreme values simultaneously (like temperature: 1.5 and top_p: 0.1), or you'll get odd outputs.
	*/
	TopP   float64 `json:"top_p"`
	Stream bool    `json:"stream"`
}

func (s *Server) SendMessage(w http.ResponseWriter, r *http.Request) {
	// Request
	maxTokens, err := strconv.Atoi(os.Getenv("MAX_TOKENS"))
	if err != nil {
		s.logger.Printf("failed to parse MAX_TOKENS: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	req := chat.ChatRequest{
		Messages: []chat.Message{
			{
				Role:    chat.MessageRoleUser,
				Content: r.URL.Query().Get("message"),
			},
		},
		Model:          chat.ModelIDLLAMA370B,
		Stream:         true,
		Temperature:    0.7,
		TopP:           0.85,
		ResponseFormat: "json",
		MaxTokens:      maxTokens,
	}

	sse, cancel, err := s.groqClient.SendMessage(r.Context(), req)
	if err != nil {
		s.logger.Printf("failed to send message: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if cancel != nil {
		defer cancel()
	}

	// get tokens and write them to the response
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	for response := range sse {
		if response.Error != nil {
			s.logger.Printf("error in SSE stream: %v", response.Error)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if response.Response.ID == "" {
			continue // skip empty responses
		}

		// Write the event to the response
		event := "data: " + response.Response.Choices[0].Delta.Content + "\n\n"
		if _, err := w.Write([]byte(event)); err != nil {
			s.logger.Printf("failed to write SSE event: %v", err)
			return
		}
		if f, ok := w.(http.Flusher); ok {
			f.Flush() // flush the buffer to ensure the client receives the data immediately
		}
	}
}

// GET /status
func (s *Server) Status(w http.ResponseWriter, r *http.Request) {
	// Respond with a simple "OK" message for liveness checks
	// TODO: check if the Groq API is reachable
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		s.logger.Printf("failed to write response: %v", err)
	}
}

// serves the index page : shows all the shares
func (s *Server) Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	s.tmpl.ExecuteTemplate(w, "index.html", nil)
}
