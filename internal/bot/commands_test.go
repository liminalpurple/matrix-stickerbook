package bot

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/liminalpurple/matrix-stickerbook/internal/config"
	"github.com/liminalpurple/matrix-stickerbook/internal/llm"
	"github.com/liminalpurple/matrix-stickerbook/internal/matrix"
	"github.com/liminalpurple/matrix-stickerbook/internal/storage"
)

// setupTestBot creates a bot with temp storage for testing
func setupTestBot(t *testing.T) (*Bot, string) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "stickerbook-cmd-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Set STICKERBOOK_CONFIG_DIR to temp directory to prevent overwriting real config
	oldConfigDir := os.Getenv("STICKERBOOK_CONFIG_DIR")
	os.Setenv("STICKERBOOK_CONFIG_DIR", tmpDir)
	t.Cleanup(func() {
		if oldConfigDir != "" {
			os.Setenv("STICKERBOOK_CONFIG_DIR", oldConfigDir)
		} else {
			os.Unsetenv("STICKERBOOK_CONFIG_DIR")
		}
	})

	matrixClient, _ := matrix.NewClient("https://matrix.org", "@test:matrix.org", "test-token")
	llmClient := llm.NewClient("test-api-key", "claude-3-haiku-20240307", 100)

	cfg := &config.Config{
		Matrix: config.MatrixConfig{
			Homeserver:  "https://matrix.org",
			UserID:      "@test:matrix.org",
			AccessToken: "test-token",
		},
		Storage: config.StorageConfig{
			DataDir: tmpDir,
		},
	}

	bot := NewBot(matrixClient, llmClient, cfg)

	return bot, tmpDir
}

// TestExecuteCommand_PackList verifies pack list command
func TestExecuteCommand_PackList(t *testing.T) {
	bot, tmpDir := setupTestBot(t)
	defer os.RemoveAll(tmpDir)
	defer bot.Stop()

	// Initially no packs - should show unsorted (0)
	result := bot.executeCommand(context.Background(), "!sticker pack list")
	if !strings.Contains(result, "unsorted (0)") {
		t.Errorf("Expected 'unsorted (0)', got: %s", result)
	}
	if !strings.Contains(result, "Create a pack") {
		t.Errorf("Expected helpful message, got: %s", result)
	}

	// Create a pack
	if err := storage.CreatePack(tmpDir, "test-pack", "Test Pack"); err != nil {
		t.Fatalf("Failed to create pack: %v", err)
	}

	// Should now show the pack and unsorted
	result = bot.executeCommand(context.Background(), "!sticker pack list")
	if !strings.Contains(result, "test-pack") {
		t.Errorf("Expected pack to be listed, got: %s", result)
	}
	if !strings.Contains(result, "unsorted (0)") {
		t.Errorf("Expected 'unsorted (0)' to always show, got: %s", result)
	}
}

// TestExecuteCommand_PackCreate verifies pack creation
func TestExecuteCommand_PackCreate(t *testing.T) {
	bot, tmpDir := setupTestBot(t)
	defer os.RemoveAll(tmpDir)
	defer bot.Stop()

	result := bot.executeCommand(context.Background(), "!sticker pack create favourites")
	if !strings.Contains(result, "✅") || !strings.Contains(result, "favourites") {
		t.Errorf("Expected success message, got: %s", result)
	}

	// Verify pack was created
	pack, err := storage.GetPack(tmpDir, "favourites")
	if err != nil {
		t.Errorf("Pack should exist: %v", err)
	}
	if pack.Name != "favourites" {
		t.Errorf("Expected pack name 'favourites', got: %s", pack.Name)
	}
}

// TestExecuteCommand_PackCreateWithSpaces verifies name sanitization
func TestExecuteCommand_PackCreateWithSpaces(t *testing.T) {
	bot, tmpDir := setupTestBot(t)
	defer os.RemoveAll(tmpDir)
	defer bot.Stop()

	result := bot.executeCommand(context.Background(), "!sticker pack create Funny Memes")
	if !strings.Contains(result, "✅") {
		t.Errorf("Expected success, got: %s", result)
	}

	// Should be converted to "funny-memes"
	_, err := storage.GetPack(tmpDir, "funny-memes")
	if err != nil {
		t.Errorf("Pack 'funny-memes' should exist: %v", err)
	}
}

// TestExecuteCommand_PackCreateUnsorted verifies "unsorted" is forbidden
func TestExecuteCommand_PackCreateUnsorted(t *testing.T) {
	bot, tmpDir := setupTestBot(t)
	defer os.RemoveAll(tmpDir)
	defer bot.Stop()

	result := bot.executeCommand(context.Background(), "!sticker pack create unsorted")
	if !strings.Contains(result, "❌") || !strings.Contains(result, "reserved") {
		t.Errorf("Expected error about reserved name, got: %s", result)
	}

	// Pack should not exist
	_, err := storage.GetPack(tmpDir, "unsorted")
	if err == nil {
		t.Error("Pack 'unsorted' should not have been created")
	}
}

// TestExecuteCommand_PackAdd verifies adding stickers to pack
func TestExecuteCommand_PackAdd(t *testing.T) {
	bot, tmpDir := setupTestBot(t)
	defer os.RemoveAll(tmpDir)
	defer bot.Stop()

	// Create pack and sticker
	if err := storage.CreatePack(tmpDir, "test-pack", "Test Pack"); err != nil {
		t.Fatalf("Failed to create pack: %v", err)
	}

	testSticker := storage.Sticker{
		ID:          "sha256:test123",
		CollectedAt: time.Now(),
		InPacks:     []string{},
	}
	if err := storage.AddSticker(tmpDir, testSticker); err != nil {
		t.Fatalf("Failed to add sticker: %v", err)
	}

	// Add sticker to pack
	result := bot.executeCommand(context.Background(), "!sticker pack add test-pack sha256:test123")
	if !strings.Contains(result, "✅") {
		t.Errorf("Expected success, got: %s", result)
	}

	// Verify it was added
	pack, _ := storage.GetPack(tmpDir, "test-pack")
	if len(pack.StickerIDs) != 1 || pack.StickerIDs[0] != "sha256:test123" {
		t.Errorf("Sticker should be in pack")
	}
}

// TestExecuteCommand_PackRemove verifies removing stickers from pack
func TestExecuteCommand_PackRemove(t *testing.T) {
	bot, tmpDir := setupTestBot(t)
	defer os.RemoveAll(tmpDir)
	defer bot.Stop()

	// Create pack and sticker, add to pack
	if err := storage.CreatePack(tmpDir, "test-pack", "Test Pack"); err != nil {
		t.Fatalf("Failed to create pack: %v", err)
	}

	testSticker := storage.Sticker{
		ID:          "sha256:test123",
		CollectedAt: time.Now(),
		InPacks:     []string{},
	}
	if err := storage.AddSticker(tmpDir, testSticker); err != nil {
		t.Fatalf("Failed to add sticker: %v", err)
	}
	if err := storage.AddToPack(tmpDir, "test-pack", []string{"sha256:test123"}); err != nil {
		t.Fatalf("Failed to add to pack: %v", err)
	}

	// Remove sticker
	result := bot.executeCommand(context.Background(), "!sticker pack remove test-pack sha256:test123")
	if !strings.Contains(result, "✅") {
		t.Errorf("Expected success, got: %s", result)
	}

	// Verify it was removed
	pack, _ := storage.GetPack(tmpDir, "test-pack")
	if len(pack.StickerIDs) != 0 {
		t.Errorf("Pack should be empty after removal")
	}
}

// TestExecuteCommand_PackShow verifies showing pack contents
func TestExecuteCommand_PackShow(t *testing.T) {
	bot, tmpDir := setupTestBot(t)
	defer os.RemoveAll(tmpDir)
	defer bot.Stop()

	// Create pack with sticker
	if err := storage.CreatePack(tmpDir, "test-pack", "Test Pack"); err != nil {
		t.Fatalf("Failed to create pack: %v", err)
	}

	testSticker := storage.Sticker{
		ID:               "test123abc",
		CollectedAt:      time.Now(),
		GeneratedAltText: "A cute sticker",
		InPacks:          []string{},
	}
	if err := storage.AddSticker(tmpDir, testSticker); err != nil {
		t.Fatalf("Failed to add sticker: %v", err)
	}
	if err := storage.AddToPack(tmpDir, "test-pack", []string{"test123abc"}); err != nil {
		t.Fatalf("Failed to add to pack: %v", err)
	}

	result := bot.executeCommand(context.Background(), "!sticker pack show test-pack")
	if !strings.Contains(result, "cute sticker") || !strings.Contains(result, "test123abc") {
		t.Errorf("Expected sticker details, got: %s", result)
	}
}

// TestExecuteCommand_ListUnsorted verifies listing unsorted stickers
func TestExecuteCommand_ListUnsorted(t *testing.T) {
	bot, tmpDir := setupTestBot(t)
	defer os.RemoveAll(tmpDir)
	defer bot.Stop()

	// Initially no stickers
	result := bot.executeCommand(context.Background(), "!sticker list unsorted")
	if !strings.Contains(result, "All stickers are organized") {
		t.Errorf("Expected organized message, got: %s", result)
	}

	// Add unsorted sticker
	testSticker := storage.Sticker{
		ID:               "sha256:test123",
		CollectedAt:      time.Now(),
		GeneratedAltText: "Unsorted sticker",
		InPacks:          []string{},
	}
	if err := storage.AddSticker(tmpDir, testSticker); err != nil {
		t.Fatalf("Failed to add sticker: %v", err)
	}

	result = bot.executeCommand(context.Background(), "!sticker list unsorted")
	if !strings.Contains(result, "Unsorted") || !strings.Contains(result, "Unsorted sticker") {
		t.Errorf("Expected unsorted sticker to be listed, got: %s", result)
	}
}

// TestExecuteCommand_InvalidCommands verifies error handling
func TestExecuteCommand_InvalidCommands(t *testing.T) {
	bot, tmpDir := setupTestBot(t)
	defer os.RemoveAll(tmpDir)
	defer bot.Stop()

	tests := []struct {
		command     string
		expectError string
		isHelp      bool
	}{
		{"!sticker", "Pack Management:", true},
		{"!sticker unknown", "Unknown command", false},
		{"!sticker pack", "No pack subcommand", false},
		{"!sticker pack unknown", "Unknown pack subcommand", false},
		{"!sticker pack create", "Usage:", false},
		{"!sticker pack add", "Usage:", false},
		{"!sticker pack add packname", "Usage:", false},
		{"!sticker list", "No list subcommand", false},
		{"!sticker list unknown", "Unknown list subcommand", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := bot.executeCommand(context.Background(), tt.command)
			if tt.isHelp {
				if !strings.Contains(result, tt.expectError) {
					t.Errorf("Expected help text with %q, got: %s", tt.expectError, result)
				}
			} else {
				if !strings.Contains(result, "❌") || !strings.Contains(result, tt.expectError) {
					t.Errorf("Expected error with %q, got: %s", tt.expectError, result)
				}
			}
		})
	}
}
