package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"smb-chatbot/internal/entity"
	"smb-chatbot/internal/usecase"
)

type conversationRepository struct {
	db *sql.DB
}

func NewConversationRepository(db *sql.DB) usecase.ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) Save(ctx context.Context, conversation *entity.Conversation) error {
	if conversation.ChatID == 0 {
		return fmt.Errorf("cannot save conversation with zero ChatID")
	}

	query := `
		INSERT INTO conversations (chat_id, user_id, state, last_interaction_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (chat_id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			state = EXCLUDED.state,
			last_interaction_at = EXCLUDED.last_interaction_at;`

	_, err := r.db.ExecContext(ctx, query, conversation.ChatID, conversation.UserID, conversation.State, conversation.LastInteractionAt)
	if err != nil {
		log.Printf("ERROR: Failed to save conversation for chat %d: %v", conversation.ChatID, err)
		return fmt.Errorf("database error saving conversation: %w", err)
	}

	log.Printf("GATEWAY (Postgres): Saved conversation state '%s' for chat %d", conversation.State, conversation.ChatID)
	return nil
}

func (r *conversationRepository) FindByChatID(ctx context.Context, chatID int64) (*entity.Conversation, error) {
	query := `SELECT chat_id, user_id, state, last_interaction_at FROM conversations WHERE chat_id = $1;`

	row := r.db.QueryRowContext(ctx, query, chatID)

	var conversation entity.Conversation
	err := row.Scan(&conversation.ChatID, &conversation.UserID, &conversation.State, &conversation.LastInteractionAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("GATEWAY (Postgres): Conversation for chat %d not found", chatID)
			log.Printf("GATEWAY (Postgres): Creating new default conversation entry for chat %d", chatID)
			newConv := entity.NewConversation(chatID, 0)
			saveErr := r.Save(ctx, newConv)
			if saveErr != nil {
				log.Printf("ERROR: Failed to save newly created default conversation for chat %d: %v", chatID, saveErr)
				return nil, fmt.Errorf("database error finding or creating conversation: %w", err)
			}
			return newConv, nil
		}
		log.Printf("ERROR: Failed to find conversation for chat %d: %v", chatID, err)
		return nil, fmt.Errorf("database error finding conversation: %w", err)
	}

	log.Printf("GATEWAY (Postgres): Found conversation state '%s' for chat %d", conversation.State, chatID)
	return &conversation, nil
}
