package persistence

type StorageType string

type Message struct {
	Role      string `json:"role"`      // Role of the message sender (e.g., "user" or "assistant")
	Content   string `json:"content"`   // Content of the Message
	Timestamp int64  `json:"timestamp"` // Timestamp of the message
}

type ConversationStore interface {
	AppendMessage(convoID string, msg Message) error
	GetRecentMessages(convoID string, limit int) ([]Message, error)
}

const (
	MemoryStorage StorageType = "memory"
)

func NewPersistence(t StorageType) ConversationStore {
	switch t {
	case MemoryStorage:
		return NewInMemoryStore()
	default:
		panic("unsupported persistence type")
	}
}
