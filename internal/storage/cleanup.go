package storage

import (
	"log"
	"time"
)

func (s *Store) StartCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			s.cleanup()
		}
	}()
}

func (s *Store) cleanup() {
	s.mu.RLock()
	var expired []string
	now := time.Now()
	for id, info := range s.files {
		if now.After(info.ExpiresAt) {
			expired = append(expired, id)
		}
	}
	s.mu.RUnlock()

	for _, id := range expired {
		log.Printf("Auto-cleanup: removing expired file %s", id)
		s.Delete(id)
	}
}
