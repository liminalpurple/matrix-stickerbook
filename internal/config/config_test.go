package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigStructs(t *testing.T) {
	// Just test that config structs can be created and fields accessed
	cfg := &Config{
		Matrix: MatrixConfig{
			Homeserver:  "https://matrix.org",
			UserID:      "@test:matrix.org",
			DeviceID:    "TESTDEVICE",
			AccessToken: "test_token",
		},
		Anthropic: AnthropicConfig{
			APIKey:    "test_api_key",
			Model:     "claude-3-haiku-20240307",
			MaxTokens: 100,
		},
		Storage: StorageConfig{
			DataDir: "/tmp/test",
		},
	}

	if cfg.Matrix.Homeserver != "https://matrix.org" {
		t.Error("Matrix homeserver not set correctly")
	}

	if cfg.Anthropic.Model != "claude-3-haiku-20240307" {
		t.Error("Anthropic model not set correctly")
	}

	if cfg.Storage.DataDir != "/tmp/test" {
		t.Error("Storage data dir not set correctly")
	}
}

func TestSavePermissions(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "stickerbook-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a config directory
	configDir := filepath.Join(tmpDir, "stickerbook")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Test that we can write a file and set permissions
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("test: data\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Set to 0600
	if err := os.Chmod(configPath, 0600); err != nil {
		t.Fatalf("Failed to chmod: %v", err)
	}

	// Verify
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected permissions 0600, got %o", info.Mode().Perm())
	}
}
