package http

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, h *ReviewController) {
	mux.HandleFunc("POST /api/message", h.handleSendMessage)
	mux.HandleFunc("GET /api/history/", h.handleGetHistory)
}
