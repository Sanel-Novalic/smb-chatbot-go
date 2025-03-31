package storage

import (
	"context"
	"testing"
	"time"

	"smb-chatbot/internal/entity"
)

func TestInMemoryReviewRepository_SaveAndFindByCustomerID(t *testing.T) {
	repo := NewInMemoryReviewRepository()
	ctx := context.Background()

	// Test data
	review1 := &entity.Review{
		ID:         "uuid-1",
		CustomerID: 101,
		ChatID:     201,
		Text:       "Great service!",
		ReceivedAt: time.Now(),
	}
	review2 := &entity.Review{
		ID:         "uuid-2",
		CustomerID: 102,
		ChatID:     202,
		Text:       "Okay.",
		ReceivedAt: time.Now(),
	}
	review3 := &entity.Review{
		ID:         "uuid-3",
		CustomerID: 101,
		ChatID:     201,
		Text:       "Very helpful.",
		ReceivedAt: time.Now().Add(time.Minute),
	}

	err := repo.Save(ctx, review1)
	if err != nil {
		t.Fatalf("Save review1 failed: %v", err)
	}
	err = repo.Save(ctx, review2)
	if err != nil {
		t.Fatalf("Save review2 failed: %v", err)
	}
	err = repo.Save(ctx, review3)
	if err != nil {
		t.Fatalf("Save review3 failed: %v", err)
	}

	// Customer 101 should have 2 reviews
	reviews101, err := repo.FindByCustomerID(ctx, 101)
	if err != nil {
		t.Fatalf("FindByCustomerID for 101 failed: %v", err)
	}
	if len(reviews101) != 2 {
		t.Fatalf("Expected 2 reviews for customer 101, got %d", len(reviews101))
	}
	// Simple check: ensure IDs are present (order might vary)
	found1, found3 := false, false
	for _, r := range reviews101 {
		if r.ID == review1.ID {
			found1 = true
		}
		if r.ID == review3.ID {
			found3 = true
		}
	}
	if !found1 || !found3 {
		t.Errorf("Did not find both review1 and review3 for customer 101. Found: %+v", reviews101)
	}

	// Customer 102 should have 1 review
	reviews102, err := repo.FindByCustomerID(ctx, 102)
	if err != nil {
		t.Fatalf("FindByCustomerID for 102 failed: %v", err)
	}
	if len(reviews102) != 1 {
		t.Fatalf("Expected 1 review for customer 102, got %d", len(reviews102))
	}
	if reviews102[0].ID != review2.ID {
		t.Errorf("Expected review for customer 102 to have ID %s, got %s", review2.ID, reviews102[0].ID)
	}

	// Customer 999 should have 0 reviews
	reviews999, err := repo.FindByCustomerID(ctx, 999)
	if err != nil {
		t.Fatalf("FindByCustomerID for 999 failed: %v", err)
	}
	if len(reviews999) != 0 {
		t.Errorf("Expected 0 reviews for customer 999, got %d", len(reviews999))
	}
}
