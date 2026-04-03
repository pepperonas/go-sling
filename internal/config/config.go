package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Auth     AuthConfig     `yaml:"auth"`
	Storage  StorageConfig  `yaml:"storage"`
	Transfer TransferConfig `yaml:"transfer"`
	UI       UIConfig       `yaml:"ui"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type AuthConfig struct {
	Pin string `yaml:"pin"`
}

type StorageConfig struct {
	DataDir       string        `yaml:"data_dir"`
	MaxUploadSize int64         `yaml:"max_upload_size"`
	RetentionHours int          `yaml:"retention_hours"`
	AutoCleanup   bool          `yaml:"auto_cleanup"`
}

type TransferConfig struct {
	ChunkSize   int `yaml:"chunk_size"`
	AckInterval int `yaml:"ack_interval"`
}

type UIConfig struct {
	DefaultTheme string `yaml:"default_theme"`
	AppName      string `yaml:"app_name"`
}

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8420,
			Host: "0.0.0.0",
		},
		Auth: AuthConfig{
			Pin: "",
		},
		Storage: StorageConfig{
			DataDir:        "./data",
			MaxUploadSize:  10 * 1024 * 1024 * 1024, // 10GB
			RetentionHours: 24,
			AutoCleanup:    true,
		},
		Transfer: TransferConfig{
			ChunkSize:   65536,
			AckInterval: 16,
		},
		UI: UIConfig{
			DefaultTheme: "dark",
			AppName:      "go-sling",
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config: %w", err)
		}
	}

	// Environment variable overrides
	if v := os.Getenv("PIN"); v != "" {
		cfg.Auth.Pin = v
	}
	if v := os.Getenv("PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = p
		}
	}
	if v := os.Getenv("DATA_DIR"); v != "" {
		cfg.Storage.DataDir = v
	}

	return cfg, nil
}

func (c *Config) WriteDefault(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

func (c *Config) RetentionDuration() time.Duration {
	return time.Duration(c.Storage.RetentionHours) * time.Hour
}
