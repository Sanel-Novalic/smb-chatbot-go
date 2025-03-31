package http

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, h *ReviewHandler) {
	mux.HandleFunc("POST /api/message", h.handleSendMessage)
	mux.HandleFunc("GET /api/history/", h.handleGetHistory)
}

func RegisterStaticFiles(mux *http.ServeMux, dir string) {
	mux.Handle("/", http.FileServer(http.Dir(dir)))
}
