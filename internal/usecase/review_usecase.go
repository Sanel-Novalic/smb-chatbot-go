package usecase

import "context"

type HandleMessageInput struct {
	ChatID   int64
	UserID   int64
	UserName string
	Text     string
}

type ReviewUseCase interface {
	HandleMessage(ctx context.Context, input HandleMessageInput) error
}
