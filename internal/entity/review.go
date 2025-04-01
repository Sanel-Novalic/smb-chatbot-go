package entity

import "time"

type Review struct {
	ID         string
	CustomerID int64 `json:"customer_id"`
	ChatID     int64 `json:"chat_id"`
	Text       string
	ReceivedAt time.Time `json:"received_at"`
}
