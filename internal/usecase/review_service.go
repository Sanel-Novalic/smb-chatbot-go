package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"smb-chatbot/internal/entity"

	"github.com/google/uuid"
	openai "github.com/sashabaranov/go-openai"
)

const chatHistoryLimit = 10

type reviewUseCase struct {
	reviewRepo   ReviewRepository
	convoRepo    ConversationRepository
	historyRepo  HistoryRepository
	messenger    MessengerClient
	openaiClient *openai.Client
}

func NewReviewUseCase(
	rr ReviewRepository,
	cr ConversationRepository,
	hr HistoryRepository,
	mc MessengerClient,
	oaiClient *openai.Client,
) ReviewUseCase {
	uc := &reviewUseCase{
		reviewRepo:   rr,
		convoRepo:    cr,
		historyRepo:  hr,
		messenger:    mc,
		openaiClient: oaiClient,
	}
	return uc
}

func (uc *reviewUseCase) HandleMessage(ctx context.Context, input HandleMessageInput) (string, error) {
	conversation, err := uc.convoRepo.FindByChatID(ctx, input.ChatID)
	if err != nil {
		log.Printf("CRITICAL ERROR: Failed to get or create conversation for chat %d: %v", input.ChatID, err)
		return "", fmt.Errorf("failed to get conversation state: %w", err)
	}

	conversation.LastInteractionAt = time.Now()

	userEntry := entity.HistoryEntry{
		IsUserMessage: true,
		Text:          input.Text,
		Timestamp:     time.Now(),
	}

	if err := uc.historyRepo.SaveHistoryEntry(ctx, input.ChatID, userEntry); err != nil {
		log.Printf("ERROR: Failed to save user message history for chat %d: %v", input.ChatID, err)
		return "", fmt.Errorf("failed to save user message: %w", err)
	}

	currentState := conversation.State

	var actionError error
	newState := currentState
	var assistantResponse string

	switch currentState {
	case entity.StateAwaitingReview:
		actionError = uc.saveReview(ctx, input, conversation)
		if actionError == nil {
			newState = entity.StateIdle
			assistantResponse, err = uc.getChatGPTResponse(ctx, input.ChatID, "The user just provided their review. Thank them for their feedback.")
			if err != nil {
				log.Printf("WARN: Failed to get ChatGPT thank you response for chat %d: %v", input.ChatID, err)
				assistantResponse = "Thanks for your feedback!"
			}
		} else {
			assistantResponse, err = uc.getChatGPTResponse(ctx, input.ChatID, "There was an error trying to save the user's review. Apologize and say we'll look into it.")
			if err != nil {
				log.Printf("WARN: Failed to get ChatGPT error response for chat %d: %v", input.ChatID, err)
				assistantResponse = "Sorry, there was an error saving your review. We'll look into it." // Fallback
			}
		}

	case entity.StateIdle:
		lowerCaseText := strings.ToLower(input.Text)
		triggerKeywords := []string{"thank", "thanks", "appreciate", "great", "awesome", "perfect", "helpful", "loved it"}
		isTrigger := false
		for _, keyword := range triggerKeywords {
			if strings.Contains(lowerCaseText, keyword) {
				isTrigger = true
				break
			}
		}

		if isTrigger {
			log.Printf("Trigger keyword found from user %d in chat %d", input.UserID, input.ChatID)
			newState = entity.StateAwaitingReview
			prompt := fmt.Sprintf("The user expressed positive sentiment ('%s'). Ask them if they would be willing to leave a quick review about their experience.", input.Text)
			assistantResponse, err = uc.getChatGPTResponse(ctx, input.ChatID, prompt)
			if err != nil {
				log.Printf("ERROR: Failed to get ChatGPT review request: %v", err)
				actionError = err
				assistantResponse = "We appreciate that! Would you mind leaving a review?"
			}
		} else {
			prompt := fmt.Sprintf("The user said: '%s'. Respond conversationally.", input.Text)
			assistantResponse, err = uc.getChatGPTResponse(ctx, input.ChatID, prompt)
			if err != nil {
				log.Printf("ERROR: Failed to get ChatGPT response: %v", err)
				actionError = err
				assistantResponse = "Sorry, I'm having trouble connecting right now."
			}
		}

	default:
		newState = entity.StateIdle
		assistantResponse, err = uc.getChatGPTResponse(ctx, input.ChatID, "My current state is unhandled. Respond generically and politely.")
		if err != nil {
			assistantResponse = "Let's start over. How can I help?"
		}
	}

	if assistantResponse != "" {
		sendErr := uc.messenger.SendMessage(ctx, input.ChatID, assistantResponse)
		if sendErr != nil {
			log.Printf("ERROR sending assistant message to chat %d: %v", input.ChatID, sendErr)
			if actionError == nil {
				actionError = fmt.Errorf("failed to send response message: %w", sendErr)
			}
		} else {
			assistantEntry := entity.HistoryEntry{
				IsUserMessage: false,
				Text:          assistantResponse,
				Timestamp:     time.Now(),
			}
			if histErr := uc.historyRepo.SaveHistoryEntry(ctx, input.ChatID, assistantEntry); histErr != nil {
				log.Printf("ERROR: Failed to save assistant message history for chat %d: %v", input.ChatID, histErr)
			}
		}
	}

	if newState != currentState {
		conversation.State = newState
		saveErr := uc.convoRepo.Save(ctx, conversation)
		if saveErr != nil {
			log.Printf("ERROR saving updated conversation state for chat %d: %v", input.ChatID, saveErr)
			if actionError == nil {
				actionError = fmt.Errorf("failed to save conversation state: %w", saveErr)
			}
		}
	} else if actionError == nil {
		saveErr := uc.convoRepo.Save(ctx, conversation)
		if saveErr != nil {
			log.Printf("ERROR saving conversation to update timestamp for chat %d: %v", input.ChatID, saveErr)
		}
	}

	return assistantResponse, actionError
}

func (uc *reviewUseCase) getChatGPTResponse(ctx context.Context, chatID int64, prompt string) (string, error) {
	history, err := uc.historyRepo.GetHistory(ctx, chatID, chatHistoryLimit)
	if err != nil {
		history = []entity.HistoryEntry{}
	}

	messages := make([]openai.ChatCompletionMessage, 0, len(history)+2) // +2 for system and current user prompt

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "You are a friendly assistant for a small business helping gather customer reviews and answer questions.",
	})

	for _, entry := range history {
		role := openai.ChatMessageRoleAssistant
		if entry.IsUserMessage {
			role = openai.ChatMessageRoleUser
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: entry.Text,
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})

	req := openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo,
		Messages: messages,
	}

	resp, err := uc.openaiClient.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Printf("ERROR: OpenAI API call failed for chat %d: %v", chatID, err)
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		log.Printf("ERROR: OpenAI returned empty response for chat %d", chatID)
		return "", fmt.Errorf("OpenAI returned empty response")
	}

	aiResponse := resp.Choices[0].Message.Content
	log.Printf("ChatGPT response for chat %d: %s", chatID, aiResponse)
	return aiResponse, nil
}

func (uc *reviewUseCase) saveReview(ctx context.Context, input HandleMessageInput, conversation *entity.Conversation) error {
	reviewID, err := uuid.NewRandom()
	if err != nil {
		log.Printf("ERROR generating UUID for review: %v", err)
		return fmt.Errorf("failed to generate review id: %w", err)
	}

	review := &entity.Review{
		ID:         reviewID.String(),
		CustomerID: input.UserID,
		ChatID:     input.ChatID,
		Text:       input.Text,
		ReceivedAt: time.Now(),
	}

	err = uc.reviewRepo.Save(ctx, review)
	if err != nil {
		log.Printf("ERROR saving review for customer %d: %v\n", input.UserID, err)
		return fmt.Errorf("failed to save review: %w", err)
	}
	log.Printf("Saved review %s from customer %d\n", review.ID, input.UserID)

	return nil
}
