package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"smb-chatbot/internal/entity"
	"smb-chatbot/internal/usecase"

	"github.com/google/uuid"
)

type reviewRepository struct {
	db *sql.DB
}

func NewReviewRepository(db *sql.DB) usecase.ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Save(ctx context.Context, review *entity.Review) error {
	if review.ID == "" {
		reviewUUID, _ := uuid.NewRandom()
		review.ID = reviewUUID.String()
		log.Printf("WARN: Generated new UUID %s for review as ID was empty", review.ID)
	}

	query := `
		INSERT INTO reviews (id, customer_id, chat_id, text, received_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			customer_id = EXCLUDED.customer_id,
			chat_id = EXCLUDED.chat_id,
			text = EXCLUDED.text,
			received_at = EXCLUDED.received_at;`

	_, err := r.db.ExecContext(ctx, query, review.ID, review.CustomerID, review.ChatID, review.Text, review.ReceivedAt)
	if err != nil {
		log.Printf("ERROR: Failed to save review %s for customer %d: %v", review.ID, review.CustomerID, err)
		return fmt.Errorf("database error saving review: %w", err)
	}

	log.Printf("GATEWAY (Postgres): Saved review %s for customer %d", review.ID, review.CustomerID)
	return nil
}
