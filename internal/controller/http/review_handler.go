package http

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	gwMessenger "smb-chatbot/internal/gateway/messenger"
	"smb-chatbot/internal/usecase"
)

type ReviewHandler struct {
	uc              usecase.ReviewUseCase            // Use case interface
	messengerClient *gwMessenger.MockMessengerClient // Concrete mock for history access
}

func NewReviewHandler(uc usecase.ReviewUseCase, mc *gwMessenger.MockMessengerClient) *ReviewHandler {
	return &ReviewHandler{
		uc:              uc,
		messengerClient: mc,
	}
}

func (h *ReviewHandler) handleSendMessage(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("Failed to decode JSON or missing fields: %v, body: %s", err, string(body))
		http.Error(w, "Invalid JSON payload. Required fields: chat_id (number), user_id (number), text (string)", http.StatusBadRequest)
		return
	}

	h.messengerClient.AddHistory(input.ChatID, true, input.Text)

	err = h.uc.HandleMessage(ctx, input)
	if err != nil {
		log.Printf("Error from use case HandleMessage: %v", err)
		http.Error(w, "Internal server error processing message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"status": "message received"}`)
}

func (h *ReviewHandler) handleGetHistory(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/history/")
	path = strings.TrimSuffix(path, "/")

	chatID, err := strconv.ParseInt(path, 10, 64)
	if err != nil || chatID == 0 {
		http.Error(w, "Invalid or missing chat_id in URL path", http.StatusBadRequest)
		return
	}
	log.Printf("HANDLER: Received GET /api/history/%d request", chatID)

	history := h.messengerClient.GetHistory(chatID)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(history)
	if err != nil {
		log.Printf("Failed to encode history response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
