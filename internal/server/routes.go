package server

import (
	"github.com/pepperonas/go-sling/internal/api"
	"github.com/pepperonas/go-sling/internal/storage"
	"github.com/pepperonas/go-sling/internal/ws"
)

func (s *Server) RegisterRoutes(store *storage.Store, hub *ws.Hub) {
	fileHandler := api.NewFileHandler(store, s.cfg, hub)
	statusHandler := api.NewStatusHandler(s, hub, store)

	// Auth
	s.mux.HandleFunc("POST /api/auth", s.auth.HandleAuth)
	s.mux.HandleFunc("GET /api/auth/status", s.auth.HandleAuthStatus)

	// Files API
	s.mux.HandleFunc("GET /api/files", fileHandler.List)
	s.mux.HandleFunc("POST /api/upload", fileHandler.Upload)
	s.mux.HandleFunc("GET /api/download/{id}", fileHandler.Download)
	s.mux.HandleFunc("DELETE /api/files/{id}", fileHandler.Delete)

	// Send to headless peer (relay)
	s.mux.HandleFunc("POST /api/send-to/{peerId}", fileHandler.SendTo)

	// Status
	s.mux.HandleFunc("GET /api/status", statusHandler.Status)

	// WebSocket
	s.mux.HandleFunc("/ws", hub.HandleWebSocket)

	// Static files (catch-all)
	s.SetupStaticFiles()
}
