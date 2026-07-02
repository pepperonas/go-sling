package config

import (
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	t.Run("default port", func(t *testing.T) {
		if cfg.Server.Port != 8420 {
			t.Errorf("Default().Server.Port = %d; want 8420", cfg.Server.Port)
		}
	})

	t.Run("default host", func(t *testing.T) {
		if cfg.Server.Host != "0.0.0.0" {
			t.Errorf("Default().Server.Host = %q; want 0.0.0.0", cfg.Server.Host)
		}
	})

	t.Run("default retention hours", func(t *testing.T) {
		if cfg.Storage.RetentionHours != 24 {
			t.Errorf("Default().Storage.RetentionHours = %d; want 24", cfg.Storage.RetentionHours)
		}
	})

	t.Run("default auto cleanup", func(t *testing.T) {
		if !cfg.Storage.AutoCleanup {
			t.Error("Default().Storage.AutoCleanup = false; want true")
		}
	})

	t.Run("default data dir", func(t *testing.T) {
		if cfg.Storage.DataDir != "./data" {
			t.Errorf("Default().Storage.DataDir = %q; want ./data", cfg.Storage.DataDir)
		}
	})

	t.Run("default app name", func(t *testing.T) {
		if cfg.UI.AppName != "go-sling" {
			t.Errorf("Default().UI.AppName = %q; want go-sling", cfg.UI.AppName)
		}
	})

	t.Run("default theme", func(t *testing.T) {
		if cfg.UI.DefaultTheme != "dark" {
			t.Errorf("Default().UI.DefaultTheme = %q; want dark", cfg.UI.DefaultTheme)
		}
	})

	t.Run("default pin is empty", func(t *testing.T) {
		if cfg.Auth.Pin != "" {
			t.Errorf("Default().Auth.Pin = %q; want empty string", cfg.Auth.Pin)
		}
	})
}

func TestAddr(t *testing.T) {
	tests := []struct {
		name string
		host string
		port int
		want string
	}{
		{
			name: "default binding",
			host: "0.0.0.0",
			port: 8420,
			want: "0.0.0.0:8420",
		},
		{
			name: "localhost",
			host: "127.0.0.1",
			port: 3000,
			want: "127.0.0.1:3000",
		},
		{
			name: "custom IP and port",
			host: "192.168.1.1",
			port: 9000,
			want: "192.168.1.1:9000",
		},
		{
			name: "port 80",
			host: "0.0.0.0",
			port: 80,
			want: "0.0.0.0:80",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Default()
			cfg.Server.Host = tc.host
			cfg.Server.Port = tc.port
			got := cfg.Addr()
			if got != tc.want {
				t.Errorf("cfg.Addr() = %q; want %q", got, tc.want)
			}
		})
	}
}

func TestRetentionDuration(t *testing.T) {
	tests := []struct {
		name           string
		retentionHours int
		want           time.Duration
	}{
		{
			name:           "24 hours (default)",
			retentionHours: 24,
			want:           24 * time.Hour,
		},
		{
			name:           "1 hour",
			retentionHours: 1,
			want:           1 * time.Hour,
		},
		{
			name:           "48 hours",
			retentionHours: 48,
			want:           48 * time.Hour,
		},
		{
			name:           "zero hours",
			retentionHours: 0,
			want:           0,
		},
		{
			name:           "one week (168 hours)",
			retentionHours: 168,
			want:           168 * time.Hour,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Default()
			cfg.Storage.RetentionHours = tc.retentionHours
			got := cfg.RetentionDuration()
			if got != tc.want {
				t.Errorf("RetentionDuration() = %v; want %v", got, tc.want)
			}
		})
	}
}
