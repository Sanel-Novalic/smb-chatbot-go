package main

import (
	"log"

	gwMessenger "smb-chatbot/internal/gateway/messenger"
	gwStorage "smb-chatbot/internal/gateway/storage"
	"smb-chatbot/internal/server"
	"smb-chatbot/internal/usecase"
)

func main() {
	log.Println("Starting SMB Chatbot...")

	reviewRepo := gwStorage.NewInMemoryReviewRepository()
	convoRepo := gwStorage.NewInMemoryConversationRepository()
	messengerClient := gwMessenger.NewMockMessengerClient()

	reviewUseCase := usecase.NewReviewUseCase(
		reviewRepo,
		convoRepo,
		messengerClient,
	)

	srv := server.NewServer(reviewUseCase, messengerClient)

	port := "8080"
	if err := srv.Start(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Println("Server stopped gracefully.")
}
