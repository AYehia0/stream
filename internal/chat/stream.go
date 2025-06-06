package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tmaxmax/go-sse"
)

// the streaming logic
type ChatStreamResponse struct {
	Response ChatResponse
	Error    error
}

// ChatStream is an interface for streaming chat responses
func (c *groqClient) SendMessage(ctx context.Context, req ChatRequest) (<-chan *ChatStreamResponse, func(), error) {

	url := fmt.Sprintf("%s/v1/chat/completions", c.BaseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	ctxWithCancel, cancel := context.WithCancel(ctx)
	httpReq, err := http.NewRequestWithContext(ctxWithCancel, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		cancel()

		return nil, nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	responseCh := make(chan *ChatStreamResponse)

	conn := sse.NewConnection(httpReq)

	// Set up the connection to handle incoming messages
	// This will handle the incoming SSE messages and send them to the response channel
	go func() {
		defer close(responseCh)

		err := conn.Connect()
		if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, context.Canceled) {
			responseCh <- &ChatStreamResponse{
				Error: fmt.Errorf("failed to connect to SSE stream: %v", err),
			}
		}
	}()

	// Callback function to handle incoming SSE events
	rm := conn.SubscribeToAll(func(e sse.Event) {
		// is done ?
		fmt.Println("Received SSE event:", e.Data)
		if strings.Contains(e.Data, "[DONE]") {
			cancel()
			return
		}

		var chatResponse ChatResponse
		err := json.Unmarshal([]byte(e.Data), &chatResponse)
		if err != nil {
			responseCh <- &ChatStreamResponse{
				Error: fmt.Errorf("failed to unmarshal chat response: %v", err),
			}
			return
		}
		responseCh <- &ChatStreamResponse{
			Response: chatResponse,
			Error:    nil,
		}
	})

	// TODO: Cleanup function to unsubscribe from the SSE connection

	return responseCh, func() {
		cancel()
		rm()
	}, nil

}
