package usecase

import "context"

type MessengerClient interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}
