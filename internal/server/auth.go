package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type AuthManager struct {
	pin      string
	dataDir  string
	sessions map[string]time.Time
	attempts map[string][]time.Time
	mu       sync.RWMutex
}

func NewAuthManager(pin string, dataDir string) *AuthManager {
	a := &AuthManager{
		pin:      pin,
		dataDir:  dataDir,
		sessions: make(map[string]time.Time),
		attempts: make(map[string][]time.Time),
	}
	a.loadSessions()
	return a
}

func (a *AuthManager) sessionsPath() string {
	return filepath.Join(a.dataDir, ".sessions.json")
}

func (a *AuthManager) saveSessions() {
	a.mu.RLock()
	defer a.mu.RUnlock()
	data, _ := json.Marshal(a.sessions)
	os.MkdirAll(a.dataDir, 0755)
	os.WriteFile(a.sessionsPath(), data, 0600)
}

func (a *AuthManager) loadSessions() {
	data, err := os.ReadFile(a.sessionsPath())
	if err != nil {
		return
	}
	json.Unmarshal(data, &a.sessions)
	// Prune expired
	now := time.Now()
	for token, expiry := range a.sessions {
		if now.After(expiry) {
			delete(a.sessions, token)
		}
	}
}

func (a *AuthManager) RequiresAuth() bool {
	return a.pin != ""
}

func (a *AuthManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.RequiresAuth() {
			next.ServeHTTP(w, r)
			return
		}

		// Allow auth endpoints through
		if r.URL.Path == "/api/auth" || r.URL.Path == "/api/auth/status" {
			next.ServeHTTP(w, r)
			return
		}

		// Allow static assets and WebSocket through (auth checked via cookie after WS upgrade)
		p := r.URL.Path
		if p == "/" || p == "/index.html" || p == "/ws" ||
			strings.HasPrefix(p, "/css/") || strings.HasPrefix(p, "/js/") || strings.HasPrefix(p, "/assets/") {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("gosling_session")
		if err != nil || !a.validSession(cookie.Value) {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *AuthManager) HandleAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	ip := r.RemoteAddr
	if !a.checkRateLimit(ip) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{"error": "too many attempts, try again later"})
		return
	}

	var req struct {
		Pin      string `json:"pin"`
		Remember bool   `json:"remember"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Pin != a.pin {
		a.recordAttempt(ip)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid PIN"})
		return
	}

	token := generateToken()
	expiry := 24 * time.Hour
	if req.Remember {
		expiry = 10 * 365 * 24 * time.Hour // forever
	}

	a.mu.Lock()
	a.sessions[token] = time.Now().Add(expiry)
	a.mu.Unlock()
	a.saveSessions()

	http.SetCookie(w, &http.Cookie{
		Name:     "gosling_session",
		Value:    token,
		Path:     "/",
		MaxAge:   int(expiry.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"expires": time.Now().Add(expiry).Unix(),
	})
}

func (a *AuthManager) HandleAuthStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"required": a.RequiresAuth(),
	})
}

func (a *AuthManager) validSession(token string) bool {
	a.mu.RLock()
	expiry, ok := a.sessions[token]
	a.mu.RUnlock()
	if !ok {
		return false
	}
	if time.Now().After(expiry) {
		a.mu.Lock()
		delete(a.sessions, token)
		a.mu.Unlock()
		return false
	}
	return true
}

func (a *AuthManager) checkRateLimit(ip string) bool {
	a.mu.RLock()
	attempts := a.attempts[ip]
	a.mu.RUnlock()

	cutoff := time.Now().Add(-1 * time.Minute)
	recent := 0
	for _, t := range attempts {
		if t.After(cutoff) {
			recent++
		}
	}
	return recent < 5
}

func (a *AuthManager) recordAttempt(ip string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.attempts[ip] = append(a.attempts[ip], time.Now())

	// Trim old attempts
	cutoff := time.Now().Add(-5 * time.Minute)
	filtered := a.attempts[ip][:0]
	for _, t := range a.attempts[ip] {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	a.attempts[ip] = filtered
}

func (a *AuthManager) CleanupSessions() {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	for token, expiry := range a.sessions {
		if now.After(expiry) {
			delete(a.sessions, token)
		}
	}
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
