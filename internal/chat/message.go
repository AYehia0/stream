package chat

type MessageRole string

// Message represents a message in the chat completion request.
type Message struct {
	Role    MessageRole `json:"role"`    // Role of the message sender (e.g., "user" or "assistant")
	Content string      `json:"content"` // Content of the message
}

const (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
)
