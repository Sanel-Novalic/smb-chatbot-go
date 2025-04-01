package main

import (
	"log"
	"os"

	gwMessenger "smb-chatbot/internal/gateway/messenger"
	gwStorage "smb-chatbot/internal/gateway/storage"
	"smb-chatbot/internal/server"
	"smb-chatbot/internal/usecase"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	log.Println("Starting SMB Chatbot...")

	err := godotenv.Load()
	if err != nil {
		log.Println("INFO: No .env file found, relying on system environment variables.")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("FATAL: DATABASE_URL environment variable not set.")
	}
	db, err := gwStorage.ConnectDB(dbURL)
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to database: %v", err)
	}
	defer func() {
		log.Println("Closing database connection...")
		if err := db.Close(); err != nil {
			log.Printf("ERROR closing database connection: %v", err)
		}
	}()

	reviewRepo := gwStorage.NewReviewRepository(db)
	convoRepo := gwStorage.NewConversationRepository(db)
	historyRepo := gwStorage.NewHistoryRepository(db)

	messengerClient := gwMessenger.NewMockMessengerClient()
	log.Println("Using Mock Messenger Client.")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("FATAL: OPENAI_API_KEY environment variable not set.")
	}
	openaiClient := openai.NewClient(apiKey)

	log.Println("OpenAI client initialized.")

	reviewUseCase := usecase.NewReviewUseCase(
		reviewRepo,
		convoRepo,
		historyRepo,
		messengerClient,
		openaiClient,
	)

	srv := server.NewServer(reviewUseCase, historyRepo, messengerClient)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Attempting to start server on port %s...", port)
	if err := srv.Start(port); err != nil {
		log.Fatalf("FATAL: Failed to start server: %v", err)
	}

	log.Println("Server stopped gracefully.")
}
