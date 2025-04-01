package http

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	gwMessenger "smb-chatbot/internal/gateway/messenger"
	"smb-chatbot/internal/usecase"
)

type ReviewController struct {
	uc              usecase.ReviewUseCase
	historyRepo     usecase.HistoryRepository
	messengerClient *gwMessenger.MockMessengerClient
}

func NewReviewController(
	uc usecase.ReviewUseCase,
	hr usecase.HistoryRepository,
	mc *gwMessenger.MockMessengerClient,
) *ReviewController {
	return &ReviewController{
		uc:              uc,
		historyRepo:     hr,
		messengerClient: mc,
	}
}

type MessageResponse struct {
	Reply string `json:"reply"`
}

func (h *ReviewController) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Println("HANDLER: Received POST /api/message request")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var input usecase.HandleMessageInput

	err = json.Unmarshal(body, &input)

	if err != nil || input.ChatID == 0 || input.UserID == 0 || input.Text == "" {
		http.Error(w, "Invalid JSON payload. Required fields: chat_id (number), user_id (number), text (string)", http.StatusBadRequest)
		return
	}

	h.messengerClient.AddHistory(input.ChatID, true, input.Text)

	botReply, err := h.uc.HandleMessage(ctx, input)
	if err != nil {
		http.Error(w, "Internal server error processing message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	responsePayload := MessageResponse{Reply: botReply}

	if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
		log.Printf("ERROR: Failed to encode response payload: %v", err)
	}
}

func (h *ReviewController) handleGetHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	path := strings.TrimPrefix(r.URL.Path, "/api/history/")
	path = strings.TrimSuffix(path, "/")

	chatID, err := strconv.ParseInt(path, 10, 64)
	if err != nil || chatID == 0 {
		http.Error(w, "Invalid or missing chat_id in URL path", http.StatusBadRequest)
		return
	}
	log.Printf("HANDLER: Received GET /api/history/%d request", chatID)

	const historyFetchLimit = 10
	history, err := h.historyRepo.GetHistory(ctx, chatID, historyFetchLimit)
	if err != nil {
		log.Printf("ERROR: Failed to get history from repository for chat %d: %v", chatID, err)
		http.Error(w, "Failed to retrieve conversation history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(history)
	if err != nil {
		log.Printf("Failed to encode history response: %v", err)
	}
}
