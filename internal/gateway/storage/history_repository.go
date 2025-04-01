package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"smb-chatbot/internal/entity"
	"smb-chatbot/internal/usecase"
)

type historyRepository struct {
	db *sql.DB
}

func NewHistoryRepository(db *sql.DB) usecase.HistoryRepository {
	return &historyRepository{db: db}
}

func (h *historyRepository) SaveHistoryEntry(ctx context.Context, chatID int64, entry entity.HistoryEntry) error {
	query := `INSERT INTO message_history (chat_id, is_user_message, text, "timestamp") VALUES ($1, $2, $3, $4);`

	_, err := h.db.ExecContext(ctx, query, chatID, entry.IsUserMessage, entry.Text, entry.Timestamp)
	if err != nil {
		log.Printf("ERROR: Failed to save history entry for chat %d: %v", chatID, err)
		return fmt.Errorf("database error saving history: %w", err)
	}
	log.Printf("GATEWAY (Postgres History Repo): Saved history entry for chat %d (user=%t)", chatID, entry.IsUserMessage)
	return nil
}

func (h *historyRepository) GetHistory(ctx context.Context, chatID int64, limit int) ([]entity.HistoryEntry, error) {
	query := `
		SELECT is_user_message, text, "timestamp"
		FROM message_history
		WHERE chat_id = $1
		ORDER BY "timestamp" DESC
		LIMIT $2;`

	rows, err := h.db.QueryContext(ctx, query, chatID, limit)
	if err != nil {
		log.Printf("ERROR: Failed to query history for chat %d: %v", chatID, err)
		return nil, fmt.Errorf("database error getting history: %w", err)
	}
	defer rows.Close()

	history := make([]entity.HistoryEntry, 0, limit)
	for rows.Next() {
		var entry entity.HistoryEntry
		err := rows.Scan(&entry.IsUserMessage, &entry.Text, &entry.Timestamp)
		if err != nil {
			log.Printf("ERROR: Failed to scan history row for chat %d: %v", chatID, err)
			return nil, fmt.Errorf("database error scanning history: %w", err)
		}
		history = append(history, entry)
	}

	if err = rows.Err(); err != nil {
		log.Printf("ERROR: Error iterating history rows for chat %d: %v", chatID, err)
		return nil, fmt.Errorf("database error iterating history: %w", err)
	}

	// The OpenAI API expects messages in chronological order.
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	log.Printf("GATEWAY (Postgres History Repo): Found %d history entries for chat %d", len(history), chatID)
	return history, nil
}
