package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/pepperonas/go-sling/internal/storage"
	"github.com/pepperonas/go-sling/internal/ws"
)

type StatusResponse struct {
	Version        string `json:"version"`
	Uptime         string `json:"uptime"`
	UptimeSeconds  int64  `json:"uptimeSeconds"`
	ConnectedPeers int    `json:"connectedPeers"`
	StorageUsed    int64  `json:"storageUsed"`
	FileCount      int    `json:"fileCount"`
	GoVersion      string `json:"goVersion"`
	GOOS           string `json:"goos"`
	GOARCH         string `json:"goarch"`
	MemAlloc       uint64 `json:"memAlloc"`
}

type serverInfo interface {
	StartTime() time.Time
}

type StatusHandler struct {
	server serverInfo
	hub    *ws.Hub
	store  *storage.Store
}

func NewStatusHandler(server serverInfo, hub *ws.Hub, store *storage.Store) *StatusHandler {
	return &StatusHandler{server: server, hub: hub, store: store}
}

func (h *StatusHandler) Status(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(h.server.StartTime())

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	resp := StatusResponse{
		Version:        "1.0.0",
		Uptime:         formatDuration(uptime),
		UptimeSeconds:  int64(uptime.Seconds()),
		ConnectedPeers: h.hub.ClientCount(),
		StorageUsed:    h.store.TotalSize(),
		FileCount:      h.store.FileCount(),
		GoVersion:      runtime.Version(),
		GOOS:           runtime.GOOS,
		GOARCH:         runtime.GOARCH,
		MemAlloc:       mem.Alloc,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}
