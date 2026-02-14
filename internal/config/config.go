// Package config handles application configuration management.
// It supports YAML files and environment variable overrides.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Matrix    MatrixConfig    `mapstructure:"matrix" yaml:"matrix"`
	Anthropic AnthropicConfig `mapstructure:"anthropic" yaml:"anthropic"`
	Storage   StorageConfig   `mapstructure:"storage" yaml:"storage"`
}

// MatrixConfig holds Matrix connection settings
type MatrixConfig struct {
	Homeserver  string `mapstructure:"homeserver" yaml:"homeserver"`
	UserID      string `mapstructure:"user_id" yaml:"user_id"`
	DeviceID    string `mapstructure:"device_id" yaml:"device_id"`
	AccessToken string `mapstructure:"access_token" yaml:"access_token"`
	NextBatch   string `mapstructure:"next_batch" yaml:"next_batch"`
}

// AnthropicConfig holds Anthropic API settings
type AnthropicConfig struct {
	APIKey    string `mapstructure:"api_key" yaml:"api_key"`
	Model     string `mapstructure:"model" yaml:"model"`
	MaxTokens int    `mapstructure:"max_tokens" yaml:"max_tokens"`
}

// StorageConfig holds storage settings
type StorageConfig struct {
	DataDir string `mapstructure:"data_dir" yaml:"data_dir"`
}

// Load reads configuration from file and environment variables
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("anthropic.model", "claude-3-haiku-20240307")
	v.SetDefault("anthropic.max_tokens", 100)

	// Determine config directory
	configDir, err := getConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to determine config directory: %w", err)
	}

	// Set default storage directory
	v.SetDefault("storage.data_dir", configDir)

	// Configure viper to read from config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir) // Will be /data in Docker (via STICKERBOOK_CONFIG_DIR env var)
	v.AddConfigPath(".")       // Current directory

	// Read config file if it exists
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is OK - we'll use defaults and env vars
	}

	// Environment variable overrides
	v.SetEnvPrefix("STICKERBOOK")
	v.AutomaticEnv()

	// Specific env var bindings
	_ = v.BindEnv("matrix.access_token", "MATRIX_ACCESS_TOKEN")
	_ = v.BindEnv("anthropic.api_key", "ANTHROPIC_API_KEY")

	// Unmarshal into config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// Save writes the current configuration to file
func Save(cfg *Config) error {
	configDir, err := getConfigDir()
	if err != nil {
		return fmt.Errorf("failed to determine config directory: %w", err)
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")

	v := viper.New()
	v.Set("matrix", cfg.Matrix)
	v.Set("anthropic", cfg.Anthropic)
	v.Set("storage", cfg.Storage)

	if err := v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Set restrictive permissions on config file (contains credentials)
	if err := os.Chmod(configPath, 0600); err != nil {
		return fmt.Errorf("failed to set config file permissions: %w", err)
	}

	return nil
}

// getConfigDir returns the configuration directory path
func getConfigDir() (string, error) {
	// Check for STICKERBOOK_CONFIG_DIR env var (Docker can set this to /data)
	if configDir := os.Getenv("STICKERBOOK_CONFIG_DIR"); configDir != "" {
		return configDir, nil
	}

	// Use XDG_CONFIG_HOME if set
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "stickerbook"), nil
	}

	// Fall back to ~/.config/stickerbook
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", "stickerbook"), nil
}

// GetConfigDir returns the configuration directory (exported for other packages)
func GetConfigDir() (string, error) {
	return getConfigDir()
}
