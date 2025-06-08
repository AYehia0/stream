package api

import (
	"encoding/json"
	"fmt"
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

type ChatRequestBody struct {
	Messages []ChatMessage `json:"messages"`
	Model    chat.ModelID  `json:"model,omitempty"`
}

// SendMessage handles the POST /chat endpoint.
//
//	@Summary		Send a message to the LLM and receive a streamed response.
//	@Description	This endpoint allows users to send a chat message to the LLM and receive a streamed response.
//	@Tags			chat
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			body	body		ChatRequestBody	true	"Chat request body"
//	@Success		200		{string}	string			"Streamed response"
//	@Failure		400		{string}	string			"Bad Request"
//	@Failure		500		{string}	string			"Internal Server Error"
//	@Router			/chat [post]
func (s *Server) SendMessage(w http.ResponseWriter, r *http.Request) { // Request
	maxTokens, err := strconv.Atoi(os.Getenv("MAX_TOKENS"))
	if err != nil {
		s.logger.Printf("failed to parse MAX_TOKENS: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// get the body from the request body
	// message should have this format: { body: [] ChatMessage{ role: "user", content: "Hello" }}
	var body ChatRequestBody
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.logger.Printf("failed to decode request body: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var model chat.ModelID
	if body.Model == "" {
		model = chat.ModelIDLLAMA38B
	} else {
		model = body.Model
	}

	req := chat.ChatRequest{
		Messages:    []chat.Message{},
		Model:       model,
		Stream:      true,
		Temperature: 0.7,
		TopP:        0.85,
		MaxTokens:   maxTokens,
	}
	// add the user messages to the request
	for _, msg := range body.Messages {
		if err = addMessageToRequest(&req, msg); err != nil {
			s.logger.Printf("failed to add message to request: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
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
			// TODO: handle internal errors accordingly
			http.Error(w, response.Error.Error(), http.StatusInternalServerError)
			return
		}

		if response.Response.ID == "" {
			continue
		}

		if content := response.Response.Choices[0].Delta.Content; content != "" {
			_, err := w.Write([]byte(content))
			if err != nil {
				s.logger.Printf("failed to write response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// add the message to the request
func addMessageToRequest(req *chat.ChatRequest, msg ChatMessage) error {
	switch msg.Role {
	case "user":
		req.Messages = append(req.Messages, chat.Message{
			Role:    chat.MessageRoleUser,
			Content: msg.Content,
		})
	case "assistant":
		req.Messages = append(req.Messages, chat.Message{
			Role:    chat.MessageRoleAssistant,
			Content: msg.Content,
		})
	case "system":
		req.Messages = append(req.Messages, chat.Message{
			Role:    chat.MessageRoleSystem,
			Content: msg.Content,
		})
	default:
		return fmt.Errorf("invalid message role: %s", msg.Role)
	}
	return nil
}

// Status handles the GET /status endpoint.
//
//	@Summary		Check the status of the server.
//	@Description	This endpoint returns the current status of the server.
//	@Tags			status
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]string	"Server status"
//	@Failure		500	{string}	string				"Internal Server Error"
//	@Router			/status [get]
func (s *Server) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{"status": "OK"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Printf("failed to write response: %v", err)
	}
}
