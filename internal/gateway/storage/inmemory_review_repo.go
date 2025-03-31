package storage

import (
	"context"
	"fmt"
	"log"
	"sync"

	"smb-chatbot/internal/entity"
	"smb-chatbot/internal/usecase"
)

type inMemoryReviewRepository struct {
	reviews    map[string]*entity.Review
	byCustomer map[int64][]string
	mu         sync.RWMutex
}

func NewInMemoryReviewRepository() usecase.ReviewRepository {
	return &inMemoryReviewRepository{
		reviews:    make(map[string]*entity.Review),
		byCustomer: make(map[int64][]string),
	}
}

func (r *inMemoryReviewRepository) Save(ctx context.Context, review *entity.Review) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if review.ID == "" {
		log.Println("ERROR: Attempted to save review with empty ID")
		return fmt.Errorf("review ID cannot be empty")
	}

	r.reviews[review.ID] = review

	found := false
	for _, id := range r.byCustomer[review.CustomerID] {
		if id == review.ID {
			found = true
			break
		}
	}
	if !found {
		r.byCustomer[review.CustomerID] = append(r.byCustomer[review.CustomerID], review.ID)
	}

	log.Printf("GATEWAY: Saved review %s for customer %d", review.ID, review.CustomerID)
	return nil
}

func (r *inMemoryReviewRepository) FindByCustomerID(ctx context.Context, customerID int64) ([]entity.Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reviewIDs, exists := r.byCustomer[customerID]
	if !exists {
		return []entity.Review{}, nil
	}

	reviews := make([]entity.Review, 0, len(reviewIDs))
	for _, id := range reviewIDs {
		if review, ok := r.reviews[id]; ok {
			reviews = append(reviews, *review)
		} else {
			log.Printf("WARN: Review ID %s found in customer index but not in main map for customer %d", id, customerID)
		}
	}

	log.Printf("GATEWAY: Found %d reviews for customer %d", len(reviews), customerID)
	return reviews, nil
}
