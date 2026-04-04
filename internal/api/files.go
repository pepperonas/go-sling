package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/pepperonas/go-sling/internal/config"
	"github.com/pepperonas/go-sling/internal/storage"
	"github.com/pepperonas/go-sling/internal/ws"
)

type FileHandler struct {
	store *storage.Store
	cfg   *config.Config
	hub   *ws.Hub
}

func NewFileHandler(store *storage.Store, cfg *config.Config, hub *ws.Hub) *FileHandler {
	return &FileHandler{store: store, cfg: cfg, hub: hub}
}

func (h *FileHandler) List(w http.ResponseWriter, r *http.Request) {
	files := h.store.List()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"files": files,
	})
}

func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > h.cfg.Storage.MaxUploadSize {
		http.Error(w, `{"error":"file too large"}`, http.StatusRequestEntityTooLarge)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.Storage.MaxUploadSize)

	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB buffer
		http.Error(w, fmt.Sprintf(`{"error":"parse error: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	var uploaded []*storage.FileInfo

	for _, fileHeaders := range r.MultipartForm.File {
		for _, fh := range fileHeaders {
			f, err := fh.Open()
			if err != nil {
				log.Printf("Error opening uploaded file: %v", err)
				continue
			}

			info, err := h.store.Save(fh.Filename, f, fh.Size)
			f.Close()
			if err != nil {
				log.Printf("Error saving file: %v", err)
				http.Error(w, fmt.Sprintf(`{"error":"save error: %s"}`, err.Error()), http.StatusInternalServerError)
				return
			}
			uploaded = append(uploaded, info)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"files": uploaded,
	})
}

func (h *FileHandler) Download(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	info, err := h.store.Get(id)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	if info.IsDir {
		if err := h.store.StreamTarGz(w, info); err != nil {
			log.Printf("Error streaming tar.gz: %v", err)
		}
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, info.Name))
	http.ServeFile(w, r, info.Path)
}

func (h *FileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.store.Delete(id); err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"ok": "deleted"})
}

func (h *FileHandler) SendTo(w http.ResponseWriter, r *http.Request) {
	peerId := r.PathValue("peerId")

	if !h.hub.IsHeadless(peerId) {
		http.Error(w, `{"error":"peer is not a headless client"}`, http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.Storage.MaxUploadSize)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"parse error: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	var uploaded []*storage.FileInfo
	for _, fileHeaders := range r.MultipartForm.File {
		for _, fh := range fileHeaders {
			f, err := fh.Open()
			if err != nil {
				log.Printf("Error opening uploaded file: %v", err)
				continue
			}
			info, err := h.store.Save(fh.Filename, f, fh.Size)
			f.Close()
			if err != nil {
				log.Printf("Error saving file: %v", err)
				http.Error(w, fmt.Sprintf(`{"error":"save error: %s"}`, err.Error()), http.StatusInternalServerError)
				return
			}
			uploaded = append(uploaded, info)
		}
	}

	// Notify headless peer about new files
	for _, info := range uploaded {
		h.hub.NotifyPeer(peerId, "file-ready", map[string]any{
			"id":   info.ID,
			"name": info.Name,
			"size": info.Size,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"files": uploaded,
	})
}
