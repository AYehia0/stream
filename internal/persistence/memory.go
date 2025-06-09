package persistence

import (
	"sync"
	"time"
)

type memoryStore struct {
	mu            sync.RWMutex
	conversations map[string][]Message
}

func NewInMemoryStore() ConversationStore {
	return &memoryStore{
		conversations: make(map[string][]Message),
	}
}

const maxMessages = 20

func (m *memoryStore) AppendMessage(convoID string, msg Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	msg.Timestamp = time.Now().Unix()
	m.conversations[convoID] = append(m.conversations[convoID], msg)

	// Trim to last 20
	if len(m.conversations[convoID]) > maxMessages {
		m.conversations[convoID] = m.conversations[convoID][len(m.conversations[convoID])-maxMessages:]
	}

	return nil
}

func (m *memoryStore) GetRecentMessages(convoID string, limit int) ([]Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	messages := m.conversations[convoID]
	if len(messages) > limit {
		return messages[len(messages)-limit:], nil
	}
	return messages, nil
}
