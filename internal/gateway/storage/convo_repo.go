package storage

import (
	"context"
	"fmt"
	"log"
	"sync"

	"smb-chatbot/internal/entity"
	"smb-chatbot/internal/usecase"
)

type inMemoryConversationRepository struct {
	conversations map[int64]*entity.Conversation
	mu            sync.RWMutex
}

func NewInMemoryConversationRepository() usecase.ConversationRepository {
	return &inMemoryConversationRepository{
		conversations: make(map[int64]*entity.Conversation),
	}
}

func (r *inMemoryConversationRepository) Save(ctx context.Context, conversation *entity.Conversation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if conversation.ChatID == 0 {
		return fmt.Errorf("cannot save conversation with zero ChatID")
	}

	convoCopy := *conversation
	r.conversations[conversation.ChatID] = &convoCopy
	log.Printf("GATEWAY: Saved conversation state '%s' for chat %d", conversation.State, conversation.ChatID)
	return nil
}

func (r *inMemoryConversationRepository) FindByChatID(ctx context.Context, chatID int64) (*entity.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conversation, exists := r.conversations[chatID]
	if !exists {
		log.Printf("GATEWAY: Conversation for chat %d not found", chatID)
		return nil, usecase.ErrConversationNotFound
	}

	convoCopy := *conversation
	log.Printf("GATEWAY: Found conversation state '%s' for chat %d", convoCopy.State, chatID)
	return &convoCopy, nil
}
