package server

import (
	"fmt"
	"log"
	"net/http"

	httpController "smb-chatbot/internal/controller/http"
	gwMessenger "smb-chatbot/internal/gateway/messenger"
	"smb-chatbot/internal/usecase"
)

type Server struct {
	reviewUseCase   usecase.ReviewUseCase
	messengerClient *gwMessenger.MockMessengerClient

	Router *http.ServeMux
}

func NewServer(uc usecase.ReviewUseCase, mc *gwMessenger.MockMessengerClient) *Server {
	s := &Server{
		reviewUseCase:   uc,
		messengerClient: mc,
		Router:          http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	reviewHandler := httpController.NewReviewHandler(s.reviewUseCase, s.messengerClient)
	httpController.RegisterRoutes(s.Router, reviewHandler)
}

func (s *Server) Start(port string) error {
	log.Printf("Starting HTTP server on port %s\n", port)
	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: s.Router,
	}
	err := httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}
