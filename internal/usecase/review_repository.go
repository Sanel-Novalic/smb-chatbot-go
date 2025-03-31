package usecase

import (
	"context"
	"smb-chatbot/internal/entity"
)

type ReviewRepository interface {
	Save(ctx context.Context, review *entity.Review) error
	FindByCustomerID(ctx context.Context, customerID int64) ([]entity.Review, error)
}
