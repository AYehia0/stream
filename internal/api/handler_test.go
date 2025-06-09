package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"stream/internal/chat"
	"stream/internal/persistence"
	"stream/pkg/logger"
	"strings"
	"sync"
	"testing"
)

type mockGroqClient struct {
	SendMessageFn func(ctx context.Context, req chat.ChatRequest) (<-chan *chat.ChatStreamResponse, func(), error)
}

func (m *mockGroqClient) SendMessage(ctx context.Context, req chat.ChatRequest) (<-chan *chat.ChatStreamResponse, func(), error) {
	return m.SendMessageFn(ctx, req)
}

func TestSendMessage_Success(t *testing.T) {
	l := logger.NewStdLogger(log.Default())
	_ = os.Setenv("MAX_TOKENS", "32")

	stream := make(chan *chat.ChatStreamResponse)
	go func() {
		stream <- &chat.ChatStreamResponse{
			Response: chat.ChatResponse{
				ID:      "some-id",
				Choices: []chat.Choice{{Delta: chat.Message{Role: "assistant", Content: "Hello"}}},
			},
			Error: nil,
		}
		close(stream)
	}()

	mockClient := &mockGroqClient{
		SendMessageFn: func(ctx context.Context, req chat.ChatRequest) (<-chan *chat.ChatStreamResponse, func(), error) {
			return stream, func() {}, nil
		},
	}

	server := &Handler{
		groqClient: mockClient,
		logger:     l,
		db:         persistence.NewInMemoryStore(),
	}

	body := ChatRequestBody{
		Messages: []ChatMessage{{Role: "user", Content: "Hello"}},
		Model:    chat.ModelIDLLAMA38B,
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	server.SendMessage(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}
	responseBody, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(responseBody), "Hello") {
		t.Fatalf("expected response to contain 'Hello', got: %s", string(responseBody))
	}
}

func TestSendMessage_InvalidJSON(t *testing.T) {
	l := logger.NewStdLogger(log.Default())
	server := &Handler{
		logger: l,
	}

	_ = os.Setenv("MAX_TOKENS", "32")

	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer([]byte("{invalid json")))
	w := httptest.NewRecorder()

	server.SendMessage(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", res.StatusCode)
	}
}

func TestSendMessage_MissingMaxTokens(t *testing.T) {
	l := logger.NewStdLogger(log.Default())
	server := &Handler{
		logger: l,
	}

	_ = os.Unsetenv("MAX_TOKENS")

	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer([]byte("{}")))
	w := httptest.NewRecorder()

	server.SendMessage(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", res.StatusCode)
	}
}

func TestSendMessage_InvalidRole(t *testing.T) {
	l := logger.NewStdLogger(log.Default())
	server := &Handler{
		logger: l,
	}

	_ = os.Setenv("MAX_TOKENS", "32")

	body := ChatRequestBody{
		Messages: []ChatMessage{{Role: "unknown", Content: "???"}},
		Model:    chat.ModelIDLLAMA38B,
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	server.SendMessage(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", res.StatusCode)
	}
}

func TestConcurrentSendMessage(t *testing.T) {
	l := logger.NewStdLogger(log.Default())
	_ = os.Setenv("MAX_TOKENS", "32")

	mockClient := &mockGroqClient{
		SendMessageFn: func(ctx context.Context, req chat.ChatRequest) (<-chan *chat.ChatStreamResponse, func(), error) {
			stream := make(chan *chat.ChatStreamResponse)
			go func() {
				defer close(stream)
				stream <- &chat.ChatStreamResponse{
					Response: chat.ChatResponse{
						ID:      "some-id",
						Choices: []chat.Choice{{Delta: chat.Message{Role: "assistant", Content: "Hello"}}},
					},
					Error: nil,
				}
			}()
			return stream, func() {}, nil
		},
	}

	server := &Handler{
		groqClient: mockClient,
		logger:     l,
		db:         persistence.NewInMemoryStore(),
	}

	var wg sync.WaitGroup
	numRequests := 10
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()
			body := ChatRequestBody{
				Messages: []ChatMessage{{Role: "user", Content: "Hello"}},
				Model:    chat.ModelIDLLAMA38B,
			}
			jsonBody, _ := json.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(jsonBody))
			w := httptest.NewRecorder()

			server.SendMessage(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				t.Errorf("expected status 200, got %d", res.StatusCode)
			}
		}()
	}

	wg.Wait()
}

func TestSendMessage_SaveConversation(t *testing.T) {
	l := logger.NewStdLogger(log.Default())
	_ = os.Setenv("MAX_TOKENS", "32")

	mockClient := &mockGroqClient{
		SendMessageFn: func(ctx context.Context, req chat.ChatRequest) (<-chan *chat.ChatStreamResponse, func(), error) {
			stream := make(chan *chat.ChatStreamResponse)
			go func() {
				defer close(stream)
				stream <- &chat.ChatStreamResponse{
					Response: chat.ChatResponse{
						ID:      "some-id",
						Choices: []chat.Choice{{Delta: chat.Message{Role: "assistant", Content: "Hello"}}},
					},
					Error: nil,
				}
			}()
			return stream, func() {}, nil
		},
	}

	server := &Handler{
		groqClient: mockClient,
		logger:     l,
		db:         persistence.NewInMemoryStore(),
	}

	body := ChatRequestBody{
		Messages: []ChatMessage{{Role: "user", Content: "Hello"}},
		Model:    chat.ModelIDLLAMA38B,
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	server.SendMessage(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}

	responseBody, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(responseBody), "Hello") {
		t.Fatalf("expected response to contain 'Hello', got: %s", string(responseBody))
	}

	// Here you would typically check if the conversation was saved in the database.
}
