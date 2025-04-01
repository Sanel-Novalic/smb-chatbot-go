package server

import (
	"fmt"
	"log"
	"net/http"

	httpController "smb-chatbot/internal/controller/http"
	gwMessenger "smb-chatbot/internal/gateway/messenger"
	"smb-chatbot/internal/usecase"

	"github.com/rs/cors"
)

type Server struct {
	reviewUseCase   usecase.ReviewUseCase
	historyRepo     usecase.HistoryRepository
	messengerClient *gwMessenger.MockMessengerClient

	Router *http.ServeMux
}

func NewServer(uc usecase.ReviewUseCase, hr usecase.HistoryRepository, mc *gwMessenger.MockMessengerClient) *Server {
	s := &Server{
		reviewUseCase:   uc,
		historyRepo:     hr,
		messengerClient: mc,
		Router:          http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	reviewHandler := httpController.NewReviewController(s.reviewUseCase, s.historyRepo, s.messengerClient)
	httpController.RegisterRoutes(s.Router, reviewHandler)
}

func (s *Server) Start(port string) error {
	log.Printf("Starting HTTP server on port %s\n", port)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            true,
	})

	handler := c.Handler(s.Router)

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}
	err := httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}
