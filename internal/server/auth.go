package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type AuthManager struct {
	pin      string
	sessions map[string]time.Time
	attempts map[string][]time.Time
	mu       sync.RWMutex
}

func NewAuthManager(pin string) *AuthManager {
	return &AuthManager{
		pin:      pin,
		sessions: make(map[string]time.Time),
		attempts: make(map[string][]time.Time),
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

		// Allow auth endpoint through
		if r.URL.Path == "/api/auth" {
			next.ServeHTTP(w, r)
			return
		}

		// Allow static assets through
		if r.URL.Path == "/" || r.URL.Path == "/index.html" ||
			len(r.URL.Path) > 4 && (r.URL.Path[:4] == "/css" || r.URL.Path[:3] == "/js" || r.URL.Path[:7] == "/assets") {
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
		expiry = 7 * 24 * time.Hour
	}

	a.mu.Lock()
	a.sessions[token] = time.Now().Add(expiry)
	a.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "gosling_session",
		Value:    token,
		Path:     "/",
		MaxAge:   int(expiry.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
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
