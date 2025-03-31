package usecase

import (
	"context"
	"errors"
	"smb-chatbot/internal/entity"
)

var ErrConversationNotFound = errors.New("conversation not found")

type ConversationRepository interface {
	Save(ctx context.Context, conversation *entity.Conversation) error
	FindByChatID(ctx context.Context, chatID int64) (*entity.Conversation, error)
}
