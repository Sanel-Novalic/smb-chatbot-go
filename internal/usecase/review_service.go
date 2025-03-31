package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"smb-chatbot/internal/entity"

	"github.com/google/uuid"
)

type reviewUseCase struct {
	reviewRepo ReviewRepository
	convoRepo  ConversationRepository
	messenger  MessengerClient

	knownUsers  map[int64]bool
	activeChats map[int64]bool
	idMu        sync.RWMutex
}

func NewReviewUseCase(
	rr ReviewRepository,
	cr ConversationRepository,
	mc MessengerClient,
) ReviewUseCase {
	uc := &reviewUseCase{
		reviewRepo:  rr,
		convoRepo:   cr,
		messenger:   mc,
		knownUsers:  make(map[int64]bool),
		activeChats: make(map[int64]bool),
	}
	// Pre-populate simulated known IDs
	uc.addKnownUser(101)
	uc.addKnownUser(2002)
	uc.addKnownUser(987)
	uc.addActiveChat(505)
	uc.addActiveChat(1001)
	return uc
}

func (uc *reviewUseCase) HandleMessage(ctx context.Context, input HandleMessageInput) error {
	if !uc.chatExists(input.ChatID) {
		return fmt.Errorf("chat %d not found", input.ChatID)
	}
	if !uc.userExists(input.UserID) {
		return fmt.Errorf("user %d not found", input.UserID)
	}

	conversation, err := uc.convoRepo.FindByChatID(ctx, input.ChatID)
	if err != nil {
		log.Printf("ERROR getting conversation for chat %d: %v", input.ChatID, err)
		return fmt.Errorf("failed to get conversation: %w", err)
	} else {
		conversation.LastInteractionAt = time.Now()
	}

	currentState := conversation.State
	log.Printf("Processing message for chat %d (User: %d, State: %s): '%s'\n", input.ChatID, input.UserID, currentState, input.Text)

	var actionError error
	newState := currentState

	switch currentState {
	case entity.StateAwaitingReview:
		actionError = uc.saveReview(ctx, input, conversation)
		if actionError == nil {
			newState = entity.StateIdle
		} // Keep state AwaitingReview if save fails? Or reset? Depends on desired retry logic.

	case entity.StateIdle:
		lowerCaseText := strings.ToLower(input.Text)
		isTrigger := strings.Contains(lowerCaseText, "thank")
		if isTrigger {
			log.Printf("Trigger keyword found from user %d in chat %d\n", input.UserID, input.ChatID)
			actionError = uc.askForReview(ctx, input, conversation)
			if actionError == nil {
				newState = entity.StateAwaitingReview
			}
		} else {
			log.Printf("Idle state, no trigger found for chat %d.", input.ChatID)
		}

	default:
		log.Printf("Unhandled state '%s' for chat %d. Resetting to Idle.", currentState, input.ChatID)
		newState = entity.StateIdle
	}

	if newState != currentState {
		conversation.State = newState
		conversation.LastInteractionAt = time.Now()
		saveErr := uc.convoRepo.Save(ctx, conversation)
		if saveErr != nil {
			log.Printf("ERROR saving updated conversation state for chat %d: %v", input.ChatID, saveErr)
			if actionError == nil {
				actionError = fmt.Errorf("failed to save conversation state: %w", saveErr)
			}
		}
	}

	return actionError
}

func (uc *reviewUseCase) askForReview(ctx context.Context, input HandleMessageInput, conversation *entity.Conversation) error {
	reviewRequestText := fmt.Sprintf("Thanks, %s! We appreciate your business. Would you mind leaving a quick review about your experience?", input.UserName)

	err := uc.messenger.SendMessage(ctx, input.ChatID, reviewRequestText)
	if err != nil {
		log.Printf("ERROR sending review request to chat %d: %v\n", input.ChatID, err)
		return fmt.Errorf("failed to send message: %w", err)
	}
	log.Printf("Sent review request to chat %d\n", input.ChatID)
	return nil
}

func (uc *reviewUseCase) saveReview(ctx context.Context, input HandleMessageInput, conversation *entity.Conversation) error {
	reviewID, _ := uuid.NewRandom()

	review := &entity.Review{
		ID:         reviewID.String(),
		CustomerID: input.UserID,
		ChatID:     input.ChatID,
		Text:       input.Text,
		ReceivedAt: time.Now(),
	}

	err := uc.reviewRepo.Save(ctx, review)
	if err != nil {
		log.Printf("ERROR saving review for customer %d: %v\n", input.UserID, err)
		return fmt.Errorf("failed to save review: %w", err)
	}
	log.Printf("Saved review %s from customer %d\n", review.ID, input.UserID)

	err = uc.messenger.SendMessage(ctx, input.ChatID, "Thanks for your feedback!")
	if err != nil {
		log.Printf("ERROR sending review confirmation to chat %d: %v\n", input.ChatID, err)
	}

	return nil
}

// Internal methods for managing simulated IDs
func (uc *reviewUseCase) addKnownUser(userID int64) {
	uc.idMu.Lock()
	defer uc.idMu.Unlock()
	uc.knownUsers[userID] = true
	log.Printf("SIM: Added known user %d\n", userID)
}
func (uc *reviewUseCase) userExists(userID int64) bool {
	uc.idMu.RLock()
	defer uc.idMu.RUnlock()
	exists := uc.knownUsers[userID]
	log.Printf("SIM: Checked user %d exists: %t\n", userID, exists)
	return exists
}
func (uc *reviewUseCase) addActiveChat(chatID int64) {
	uc.idMu.Lock()
	defer uc.idMu.Unlock()
	uc.activeChats[chatID] = true
	log.Printf("SIM: Added active chat %d\n", chatID)
}
func (uc *reviewUseCase) chatExists(chatID int64) bool {
	uc.idMu.RLock()
	defer uc.idMu.RUnlock()
	exists := uc.activeChats[chatID]
	log.Printf("SIM: Checked chat %d exists: %t\n", chatID, exists)
	return exists
}
