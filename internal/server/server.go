package server

import (
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"time"

	qrcode "github.com/skip2/go-qrcode"

	"github.com/pepperonas/go-sling/internal/config"
)

type Server struct {
	cfg     *config.Config
	auth    *AuthManager
	mux     *http.ServeMux
	webFS   fs.FS
	startAt time.Time
}

func New(cfg *config.Config, webFS fs.FS) *Server {
	return &Server{
		cfg:     cfg,
		auth:    NewAuthManager(cfg.Auth.Pin),
		mux:     http.NewServeMux(),
		webFS:   webFS,
		startAt: time.Now(),
	}
}

func (s *Server) Auth() *AuthManager {
	return s.auth
}

func (s *Server) Mux() *http.ServeMux {
	return s.mux
}

func (s *Server) Config() *config.Config {
	return s.cfg
}

func (s *Server) StartTime() time.Time {
	return s.startAt
}

func (s *Server) SetupStaticFiles() {
	s.mux.Handle("/", http.FileServer(http.FS(s.webFS)))
}

func (s *Server) Handler() http.Handler {
	var handler http.Handler = s.mux

	// Auth middleware
	handler = s.auth.Middleware(handler)

	// Logging middleware
	handler = loggingMiddleware(handler)

	// CORS for local network
	handler = corsMiddleware(handler)

	return handler
}

func (s *Server) Start() error {
	addr := s.cfg.Addr()

	if !s.auth.RequiresAuth() {
		log.Println("WARNING: No PIN configured. Access is unrestricted.")
	}

	printBanner(addr)

	srv := &http.Server{
		Addr:         addr,
		Handler:      s.Handler(),
		ReadTimeout:  0, // no timeout for large uploads
		WriteTimeout: 0, // no timeout for large downloads
		IdleTimeout:  120 * time.Second,
	}

	return srv.ListenAndServe()
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.status, time.Since(start).Round(time.Millisecond))
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func printBanner(addr string) {
	ip := getLocalIP()
	fmt.Println()
	fmt.Println("  ┌─────────────────────────────────────┐")
	fmt.Println("  │           go-sling v1.0.0            │")
	fmt.Println("  │       LAN File Transfer Server       │")
	fmt.Println("  ├─────────────────────────────────────┤")
	fmt.Printf("  │  Local:   http://%-19s │\n", addr)

	networkURL := ""
	if ip != "" {
		// Extract port from addr
		_, port, _ := net.SplitHostPort(addr)
		networkURL = fmt.Sprintf("http://%s:%s", ip, port)
		fmt.Printf("  │  Network: %-27s │\n", networkURL)
	}
	fmt.Println("  └─────────────────────────────────────┘")

	if networkURL != "" {
		qr, err := qrcode.New(networkURL, qrcode.Medium)
		if err == nil {
			fmt.Println()
			fmt.Println("  Scan to open:")
			fmt.Println(qr.ToSmallString(false))
		}
	}
	fmt.Println()
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String()
		}
	}
	return ""
}
