package messenger

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"smb-chatbot/internal/entity"
)

type MockMessengerClient struct {
	mu           sync.RWMutex
	SentMessages map[int64][]string
	History      map[int64][]entity.HistoryEntry
}

func NewMockMessengerClient() *MockMessengerClient {
	return &MockMessengerClient{
		SentMessages: make(map[int64][]string),
		History:      make(map[int64][]entity.HistoryEntry),
	}
}

func (m *MockMessengerClient) SendMessage(ctx context.Context, chatID int64, text string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fmt.Printf("MOCK MESSENGER: Attempting to send message to chat %d: %s\n", chatID, text)
	m.SentMessages[chatID] = append(m.SentMessages[chatID], text)

	entry := entity.HistoryEntry{
		IsUserMessage: false,
		Text:          text,
		Timestamp:     time.Now(),
	}
	m.History[chatID] = append(m.History[chatID], entry)

	return nil
}

func (m *MockMessengerClient) AddHistory(chatID int64, isUser bool, text string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry := entity.HistoryEntry{
		IsUserMessage: isUser,
		Text:          text,
		Timestamp:     time.Now(),
	}
	m.History[chatID] = append(m.History[chatID], entry)
	log.Printf("MOCK MESSENGER: Added history for chat %d (user=%t): %s\n", chatID, isUser, text)
}

func (m *MockMessengerClient) GetHistory(chatID int64) []entity.HistoryEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	historyCopy := make([]entity.HistoryEntry, len(m.History[chatID]))
	copy(historyCopy, m.History[chatID])
	return historyCopy
}
