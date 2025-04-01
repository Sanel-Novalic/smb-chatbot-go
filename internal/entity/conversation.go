package entity

import "time"

type Conversation struct {
	ChatID            int64
	UserID            int64
	State             string
	LastInteractionAt time.Time
}

type HistoryEntry struct {
	IsUserMessage bool      `json:"is_user_message"`
	Text          string    `json:"text"`
	Timestamp     time.Time `json:"timestamp"`
}

const (
	StateIdle           = "Idle"
	StateAwaitingReview = "AwaitingReview"
)

func NewConversation(chatID, userID int64) *Conversation {
	return &Conversation{
		ChatID:            chatID,
		UserID:            userID,
		State:             StateIdle,
		LastInteractionAt: time.Now(),
	}
}
