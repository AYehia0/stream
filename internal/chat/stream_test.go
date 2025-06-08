package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// fakeSSEHandler simulates a Groq-like SSE endpoint.
func fakeSSEHandler(t *testing.T, chunks []string, delay time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected http.ResponseWriter to be a Flusher")
		}

		for _, chunk := range chunks {
			fmt.Fprintf(w, "data: %s\n\n", chunk)
			flusher.Flush()
			time.Sleep(delay)
		}
	}
}

func TestSendMessage_StreamSuccess(t *testing.T) {
	mockResponse := ChatResponse{
		ID: "chatcmpl-123",
		Choices: []Choice{{
			Delta: Message{Role: MessageRoleAssistant, Content: "Hello"},
		}},
	}

	// Prepare test server
	chunks := []string{}
	for i := 0; i < 3; i++ {
		respBytes, _ := json.Marshal(mockResponse)
		chunks = append(chunks, string(respBytes))
	}
	chunks = append(chunks, "[DONE]")

	server := httptest.NewServer(fakeSSEHandler(t, chunks, 10*time.Millisecond))
	defer server.Close()

	client := &groqClient{
		BaseURL: server.URL,
		APIKey:  "fake-key",
	}

	req := ChatRequest{
		Model:  ModelIDLLAMA370B,
		Stream: true,
		Messages: []Message{
			{Role: MessageRoleUser, Content: "Hello"},
		},
	}

	ctx := context.Background()
	stream, cancel, err := client.SendMessage(ctx, req)
	defer cancel()
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}

	var responses []ChatResponse
	for msg := range stream {
		if msg.Error != nil {
			t.Errorf("unexpected error: %v", msg.Error)
			continue
		}
		responses = append(responses, msg.Response)
	}

	if len(responses) != 3 {
		t.Errorf("expected 3 responses, got %d", len(responses))
	}

	if responses[0].Choices[0].Delta.Content != "Hello" {
		t.Errorf("expected content 'Hello', got '%s'", responses[0].Choices[0].Delta.Content)
	}
}

func TestSendMessage_StreamMalformedJSON(t *testing.T) {
	chunks := []string{`{"invalid_json": `, "[DONE]"}
	server := httptest.NewServer(fakeSSEHandler(t, chunks, 0))
	defer server.Close()

	client := &groqClient{
		BaseURL: server.URL,
		APIKey:  "fake-key",
	}

	req := ChatRequest{
		Model:  ModelIDLLAMA370B,
		Stream: true,
		Messages: []Message{
			{Role: MessageRoleUser, Content: "Hi"},
		},
	}

	stream, cancel, err := client.SendMessage(context.Background(), req)
	defer cancel()
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}

	var gotError bool
	for msg := range stream {
		if msg.Error != nil {
			gotError = true
			if !strings.Contains(msg.Error.Error(), "unmarshal") {
				t.Errorf("expected unmarshal error, got: %v", msg.Error)
			}
		}
	}

	if !gotError {
		t.Error("expected error but got none")
	}
}

func TestSendMessage_StreamEOF(t *testing.T) {
	chunks := []string{"[DONE]"}
	server := httptest.NewServer(fakeSSEHandler(t, chunks, 0))
	defer server.Close()

	client := &groqClient{
		BaseURL: server.URL,
		APIKey:  "fake-key",
	}

	req := ChatRequest{
		Model:  ModelIDLLAMA370B,
		Stream: true,
		Messages: []Message{
			{Role: MessageRoleUser, Content: "Test"},
		},
	}

	ctx := context.Background()
	stream, cancel, err := client.SendMessage(ctx, req)
	defer cancel()
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}

	var responses []ChatResponse
	for msg := range stream {
		if msg.Error != nil {
			t.Errorf("unexpected error: %v", msg.Error)
			continue
		}
		responses = append(responses, msg.Response)
	}

	if len(responses) != 0 {
		t.Errorf("expected 0 responses, got %d", len(responses))
	}
}
