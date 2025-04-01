package usecase

import (
	"context"
	"smb-chatbot/internal/entity"
)

type HistoryRepository interface {
	SaveHistoryEntry(ctx context.Context, chatID int64, entry entity.HistoryEntry) error
	GetHistory(ctx context.Context, chatID int64, limit int) ([]entity.HistoryEntry, error)
}
