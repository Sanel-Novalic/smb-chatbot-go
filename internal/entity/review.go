package entity

import "time"

type Review struct {
	ID         string
	CustomerID int64
	ChatID     int64
	Text       string
	ReceivedAt time.Time
}
