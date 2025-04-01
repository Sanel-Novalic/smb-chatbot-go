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

func (uc *reviewUseCase) getChatGPTAnalysis(ctx context.Context, chatID int64, prompt string) (string, error) {
	history, err := uc.historyRepo.GetHistory(ctx, chatID, chatHistoryLimit) // Use the same limit const
	if err != nil {
		log.Printf("WARN (Analysis): Failed to get history for chat %d: %v. Proceeding without history.", chatID, err)
		history = []entity.HistoryEntry{}
	}

	messages := make([]openai.ChatCompletionMessage, 0, len(history)+2)
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "You are an AI analyzing conversation context.",
	})
	for _, entry := range history {
		role := openai.ChatMessageRoleAssistant
		if entry.IsUserMessage {
			role = openai.ChatMessageRoleUser
		}
		messages = append(messages, openai.ChatCompletionMessage{Role: role, Content: entry.Text})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})

	req := openai.ChatCompletionRequest{
		Model:       openai.GPT3Dot5Turbo,
		Messages:    messages,
		MaxTokens:   10,
		Temperature: 0.0,
	}

	resp, err := uc.openaiClient.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Printf("ERROR (Analysis): OpenAI API call failed for chat %d: %v", chatID, err)
		return "", fmt.Errorf("OpenAI API error during analysis: %w", err)
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		log.Printf("ERROR (Analysis): OpenAI returned empty response for chat %d", chatID)
		return "", fmt.Errorf("OpenAI returned empty analysis response")
	}

	analysisResult := strings.TrimSpace(strings.ToUpper(resp.Choices[0].Message.Content))
	log.Printf("ChatGPT analysis result for chat %d: %s", chatID, analysisResult)
	return analysisResult, nil
}

func (uc *reviewUseCase) HandleMessage(ctx context.Context, input HandleMessageInput) (string, error) {
	conversation, err := uc.convoRepo.FindByChatID(ctx, input.ChatID)
	if err != nil {
		return "", fmt.Errorf("failed to get conversation state: %w", err)
	}
	if conversation.UserID == 0 && input.UserID != 0 {
		conversation.UserID = input.UserID
	}
	conversation.LastInteractionAt = time.Now()

	currentState := conversation.State

	var actionError error
	newState := currentState
	var assistantResponse string
	saveUserMessage := true

	switch currentState {
	case entity.StateIdle:
		analysisPrompt := fmt.Sprintf(
			"Analyze the sentiment and context of the following user message. "+
				"Is the user expressing definite gratitude, concluding satisfaction, or clearly ending the conversation positively? "+
				"Respond with only 'YES' or 'NO'. Message: '%s'", input.Text,
		)
		triggerAnalysis, analysisErr := uc.getChatGPTAnalysis(ctx, input.ChatID, analysisPrompt)
		saveUserMessage = false

		if analysisErr != nil {
			assistantResponse, err = uc.getChatGPTResponse(ctx, input.ChatID, fmt.Sprintf("The user said: '%s'. Respond conversationally.", input.Text))
			if err != nil {
				actionError = err
				assistantResponse = "Sorry, I couldn't process that."
			}
			saveUserMessage = true
		} else if triggerAnalysis == "YES" {
			newState = entity.StateAwaitingReview
			reviewRequestPrompt := "The user's last message indicated satisfaction. Ask them politely if they would be willing to leave a quick review about their experience."
			assistantResponse, err = uc.getChatGPTResponse(ctx, input.ChatID, reviewRequestPrompt)
			if err != nil {
				actionError = err
				assistantResponse = "We appreciate that! Would you mind leaving a review?"
			}
			saveUserMessage = true
		} else {
			normalReplyPrompt := fmt.Sprintf("The user said: '%s'. Respond conversationally.", input.Text)
			assistantResponse, err = uc.getChatGPTResponse(ctx, input.ChatID, normalReplyPrompt)
			if err != nil {
				actionError = err
				assistantResponse = "Sorry, I couldn't process that."
			}
			saveUserMessage = true
		}

	case entity.StateAwaitingReview:
		analysisPrompt := fmt.Sprintf(
			"Analyze the following user message. Does it appear to be a genuine attempt at providing review feedback "+
				"(positive, negative, or neutral), rather than asking a question, changing the subject, or refusing? "+
				"Respond with only 'YES' or 'NO'. Message: '%s'", input.Text,
		)
		reviewAnalysis, analysisErr := uc.getChatGPTAnalysis(ctx, input.ChatID, analysisPrompt)
		saveUserMessage = true

		if analysisErr != nil {
			repromptPrompt := "There was an issue processing your previous message. Could you please provide your feedback on the experience?"
			assistantResponse, err = uc.getChatGPTResponse(ctx, input.ChatID, repromptPrompt)
			if err != nil {
				actionError = err
				assistantResponse = "Could you please provide your review?"
			}
		} else if reviewAnalysis == "YES" {
			log.Printf("ChatGPT analysis suggests input is a review for chat %d", input.ChatID)
			actionError = uc.saveReview(ctx, input, conversation)
			if actionError == nil {
				newState = entity.StateIdle
				thankPrompt := "The user provided a review. Thank them for their feedback."
				assistantResponse, err = uc.getChatGPTResponse(ctx, input.ChatID, thankPrompt)
				if err != nil {
					actionError = err
					assistantResponse = "Thanks for your feedback!"
				}
			} else {
				errorPrompt := "There was an error saving the user's review. Apologize and say we'll look into it."
				assistantResponse, err = uc.getChatGPTResponse(ctx, input.ChatID, errorPrompt)
				if err != nil {
					actionError = err
					assistantResponse = "Sorry, there was an error saving your review."
				}
				newState = entity.StateIdle
			}
		} else {
			repromptPrompt := "That doesn't seem like review feedback. Could you please share your thoughts on your experience with us? If you don't want to leave feedback right now, just let me know."
			assistantResponse, err = uc.getChatGPTResponse(ctx, input.ChatID, repromptPrompt)
			if err != nil {
				actionError = err
				assistantResponse = "Could you please provide your review?"
			}
		}

	default:
		log.Printf("Unhandled state '%s' for chat %d. Resetting to Idle.", currentState, input.ChatID)
		newState = entity.StateIdle
		assistantResponse, err = uc.getChatGPTResponse(ctx, input.ChatID, "My current state is unhandled. Respond generically.")
		if err != nil {
			assistantResponse = "Let's start over."
		}
		saveUserMessage = true
	}

	if saveUserMessage {
		userEntry := entity.HistoryEntry{
			IsUserMessage: true,
			Text:          input.Text,
			Timestamp:     time.Now(),
		}
		if histErr := uc.historyRepo.SaveHistoryEntry(ctx, input.ChatID, userEntry); histErr != nil {
			log.Printf("ERROR: Failed to save user message history for chat %d: %v", input.ChatID, histErr)
			if actionError == nil {
				actionError = fmt.Errorf("failed to save user message: %w", histErr)
			}
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
			if actionError == nil {
				actionError = fmt.Errorf("failed to save conversation state: %w", saveErr)
			}
		}
	} else if actionError == nil {
		saveErr := uc.convoRepo.Save(ctx, conversation)
		if saveErr != nil {
			log.Printf("ERROR saving conversation timestamp update for chat %d: %v", input.ChatID, saveErr)
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
