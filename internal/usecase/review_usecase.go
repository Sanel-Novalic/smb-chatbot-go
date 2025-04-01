package usecase

import "context"

type HandleMessageInput struct {
	ChatID   int64  `json:"chat_id"`
	UserID   int64  `json:"user_id"`
	UserName string `json:"user_name"`
	Text     string
}

type ReviewUseCase interface {
	HandleMessage(ctx context.Context, input HandleMessageInput) (string, error)
}
