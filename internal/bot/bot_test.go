package bot

import (
	"os"
	"testing"

	"github.com/liminalpurple/matrix-stickerbook/internal/config"
	"github.com/liminalpurple/matrix-stickerbook/internal/llm"
	"github.com/liminalpurple/matrix-stickerbook/internal/matrix"
	"maunium.net/go/mautrix/event"
)

// getTestStorageDir returns the storage directory for tests (env var or default)
func getTestStorageDir() string {
	if dir := os.Getenv("STICKERBOOK_TEST_STORAGE"); dir != "" {
		return dir
	}
	return "/tmp/test-storage"
}

// testConfig creates a minimal config for testing
func testConfig(storageDir string) *config.Config {
	return &config.Config{
		Matrix: config.MatrixConfig{
			Homeserver:  "https://matrix.org",
			UserID:      "@test:matrix.org",
			AccessToken: "test-token",
		},
		Storage: config.StorageConfig{
			DataDir: storageDir,
		},
	}
}

// setupTestEnv sets up test environment to prevent overwriting real config
func setupTestEnv(t *testing.T) func() {
	t.Helper()

	// Set STICKERBOOK_CONFIG_DIR to temp directory
	oldConfigDir := os.Getenv("STICKERBOOK_CONFIG_DIR")
	if err := os.Setenv("STICKERBOOK_CONFIG_DIR", getTestStorageDir()); err != nil {
		t.Fatalf("Failed to set test env: %v", err)
	}

	// Return cleanup function
	return func() {
		if oldConfigDir != "" {
			_ = os.Setenv("STICKERBOOK_CONFIG_DIR", oldConfigDir)
		} else {
			_ = os.Unsetenv("STICKERBOOK_CONFIG_DIR")
		}
	}
}

// TestNewBot verifies bot creation
func TestNewBot(t *testing.T) {
	defer setupTestEnv(t)()
	// Create minimal clients for testing
	matrixClient, err := matrix.NewClient("https://matrix.org", "@test:matrix.org", "test-token")
	if err != nil {
		t.Fatalf("Failed to create matrix client: %v", err)
	}

	llmClient := llm.NewClient("test-api-key", "claude-3-haiku-20240307", 100)

	bot := NewBot(matrixClient, llmClient, testConfig(getTestStorageDir()))

	if bot == nil {
		t.Fatal("Expected bot to be created")
	}

	if bot.client != matrixClient {
		t.Error("Expected matrix client to be set")
	}

	if bot.llmClient != llmClient {
		t.Error("Expected LLM client to be set")
	}

	if bot.storageDir != "/tmp/test-storage" {
		t.Errorf("Expected storage dir /tmp/test-storage, got %s", bot.storageDir)
	}

	if bot.ctx == nil {
		t.Error("Expected context to be initialized")
	}

	if bot.cancel == nil {
		t.Error("Expected cancel function to be initialized")
	}

	if bot.syncer == nil {
		t.Error("Expected syncer to be initialized")
	}

	// Clean up
	bot.Stop()
}

// TestValidCommands verifies command recognition
func TestValidCommands(t *testing.T) {
	defer setupTestEnv(t)()

	tests := []struct {
		command string
		valid   bool
	}{
		{"!yoink", true},
		{"!nom", true},
		{"!grab", true},
		{"!invalid", false},
		{"yoink", false},
		{"", false},
		{"👍", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := validCommands[tt.command]
			if result != tt.valid {
				t.Errorf("Command %q: expected valid=%v, got %v", tt.command, tt.valid, result)
			}
		})
	}
}

// TestBotStop verifies graceful shutdown
func TestBotStop(t *testing.T) {
	defer setupTestEnv(t)()
	matrixClient, err := matrix.NewClient("https://matrix.org", "@test:matrix.org", "test-token")
	if err != nil {
		t.Fatalf("Failed to create matrix client: %v", err)
	}

	llmClient := llm.NewClient("test-api-key", "claude-3-haiku-20240307", 100)
	bot := NewBot(matrixClient, llmClient, testConfig(getTestStorageDir()))

	// Context should be active
	select {
	case <-bot.ctx.Done():
		t.Error("Context should not be cancelled initially")
	default:
		// Expected
	}

	// Stop the bot
	bot.Stop()

	// Context should now be cancelled
	select {
	case <-bot.ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled after Stop()")
	}
}

// TestExtractImageData_Sticker verifies extracting data from m.sticker events
func TestExtractImageData_Sticker(t *testing.T) {
	defer setupTestEnv(t)()
	matrixClient, _ := matrix.NewClient("https://matrix.org", "@test:matrix.org", "test-token")
	llmClient := llm.NewClient("test-api-key", "claude-3-haiku-20240307", 100)
	bot := NewBot(matrixClient, llmClient, testConfig(getTestStorageDir()))
	defer bot.Stop()

	// Create a mock sticker event
	evt := &event.Event{
		Type: event.EventSticker,
		Content: event.Content{
			Parsed: &event.MessageEventContent{
				URL:  "mxc://matrix.org/test123",
				Body: "Cool sticker",
			},
		},
	}

	mxcURI, body, err := bot.extractImageData(evt)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if string(mxcURI) != "mxc://matrix.org/test123" {
		t.Errorf("Expected MXC URI 'mxc://matrix.org/test123', got %s", mxcURI)
	}

	if body != "Cool sticker" {
		t.Errorf("Expected body 'Cool sticker', got %s", body)
	}
}

// TestExtractImageData_StickerRawContent verifies sticker with raw content (unparsed)
func TestExtractImageData_StickerRawContent(t *testing.T) {
	defer setupTestEnv(t)()
	matrixClient, _ := matrix.NewClient("https://matrix.org", "@test:matrix.org", "test-token")
	llmClient := llm.NewClient("test-api-key", "claude-3-haiku-20240307", 100)
	bot := NewBot(matrixClient, llmClient, testConfig(getTestStorageDir()))
	defer bot.Stop()

	// Create a sticker event with raw content (not parsed)
	evt := &event.Event{
		Type: event.EventSticker,
		Content: event.Content{
			Raw: map[string]interface{}{
				"url":  "mxc://matrix.org/rawtest456",
				"body": "Raw sticker",
			},
		},
	}

	mxcURI, body, err := bot.extractImageData(evt)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if string(mxcURI) != "mxc://matrix.org/rawtest456" {
		t.Errorf("Expected MXC URI 'mxc://matrix.org/rawtest456', got %s", mxcURI)
	}

	if body != "Raw sticker" {
		t.Errorf("Expected body 'Raw sticker', got %s", body)
	}
}

// TestExtractImageData_ImageMessage verifies extracting data from m.room.message with m.image
func TestExtractImageData_ImageMessage(t *testing.T) {
	defer setupTestEnv(t)()
	matrixClient, _ := matrix.NewClient("https://matrix.org", "@test:matrix.org", "test-token")
	llmClient := llm.NewClient("test-api-key", "claude-3-haiku-20240307", 100)
	bot := NewBot(matrixClient, llmClient, testConfig(getTestStorageDir()))
	defer bot.Stop()

	// Create a mock image message event
	evt := &event.Event{
		Type: event.EventMessage,
		Content: event.Content{
			Parsed: &event.MessageEventContent{
				MsgType: event.MsgImage,
				URL:     "mxc://matrix.org/image456",
				Body:    "Screenshot",
			},
		},
	}

	mxcURI, body, err := bot.extractImageData(evt)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if string(mxcURI) != "mxc://matrix.org/image456" {
		t.Errorf("Expected MXC URI 'mxc://matrix.org/image456', got %s", mxcURI)
	}

	if body != "Screenshot" {
		t.Errorf("Expected body 'Screenshot', got %s", body)
	}
}

// TestExtractImageData_ImageMessageRawContent verifies m.room.message with raw content
func TestExtractImageData_ImageMessageRawContent(t *testing.T) {
	defer setupTestEnv(t)()
	matrixClient, _ := matrix.NewClient("https://matrix.org", "@test:matrix.org", "test-token")
	llmClient := llm.NewClient("test-api-key", "claude-3-haiku-20240307", 100)
	bot := NewBot(matrixClient, llmClient, testConfig(getTestStorageDir()))
	defer bot.Stop()

	// Create an image message event with raw content (not parsed)
	evt := &event.Event{
		Type: event.EventMessage,
		Content: event.Content{
			Raw: map[string]interface{}{
				"msgtype": "m.image",
				"url":     "mxc://matrix.org/rawimage789",
				"body":    "Raw image message",
			},
		},
	}

	mxcURI, body, err := bot.extractImageData(evt)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if string(mxcURI) != "mxc://matrix.org/rawimage789" {
		t.Errorf("Expected MXC URI 'mxc://matrix.org/rawimage789', got %s", mxcURI)
	}

	if body != "Raw image message" {
		t.Errorf("Expected body 'Raw image message', got %s", body)
	}
}

// TestExtractImageData_TextMessage verifies error when message is not an image
func TestExtractImageData_TextMessage(t *testing.T) {
	defer setupTestEnv(t)()
	matrixClient, _ := matrix.NewClient("https://matrix.org", "@test:matrix.org", "test-token")
	llmClient := llm.NewClient("test-api-key", "claude-3-haiku-20240307", 100)
	bot := NewBot(matrixClient, llmClient, testConfig(getTestStorageDir()))
	defer bot.Stop()

	// Create a text message event (not an image)
	evt := &event.Event{
		Type: event.EventMessage,
		Content: event.Content{
			Parsed: &event.MessageEventContent{
				MsgType: event.MsgText,
				Body:    "Just a text message",
			},
		},
	}

	_, _, err := bot.extractImageData(evt)
	if err == nil {
		t.Error("Expected error when extracting from text message")
	}

	expectedError := "message is not an image (msgtype=m.text)"
	if err.Error() != expectedError {
		t.Errorf("Expected error message %q, got %q", expectedError, err.Error())
	}
}

// TestExtractImageData_UnsupportedEventType verifies error on unsupported event types
func TestExtractImageData_UnsupportedEventType(t *testing.T) {
	defer setupTestEnv(t)()
	matrixClient, _ := matrix.NewClient("https://matrix.org", "@test:matrix.org", "test-token")
	llmClient := llm.NewClient("test-api-key", "claude-3-haiku-20240307", 100)
	bot := NewBot(matrixClient, llmClient, testConfig(getTestStorageDir()))
	defer bot.Stop()

	tests := []struct {
		name      string
		eventType event.Type
	}{
		{"m.room.member", event.StateMember},
		{"m.room.create", event.StateCreate},
		{"m.reaction", event.EventReaction},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := &event.Event{
				Type: tt.eventType,
				Content: event.Content{
					Parsed: &event.MessageEventContent{},
				},
			}

			_, _, err := bot.extractImageData(evt)
			if err == nil {
				t.Errorf("Expected error for event type %s", tt.eventType.Type)
			}
		})
	}
}

// TestExtractImageData_InvalidContent verifies error handling for malformed content
func TestExtractImageData_InvalidContent(t *testing.T) {
	defer setupTestEnv(t)()
	matrixClient, _ := matrix.NewClient("https://matrix.org", "@test:matrix.org", "test-token")
	llmClient := llm.NewClient("test-api-key", "claude-3-haiku-20240307", 100)
	bot := NewBot(matrixClient, llmClient, testConfig(getTestStorageDir()))
	defer bot.Stop()

	// Create sticker event with wrong content type
	evt := &event.Event{
		Type: event.EventSticker,
		Content: event.Content{
			Parsed: "not a MessageEventContent", // Wrong type
		},
	}

	_, _, err := bot.extractImageData(evt)
	if err == nil {
		t.Error("Expected error when content is not MessageEventContent")
	}
}

// TestExtractImageData_VideoMessage verifies video messages are rejected
func TestExtractImageData_VideoMessage(t *testing.T) {
	defer setupTestEnv(t)()
	matrixClient, _ := matrix.NewClient("https://matrix.org", "@test:matrix.org", "test-token")
	llmClient := llm.NewClient("test-api-key", "claude-3-haiku-20240307", 100)
	bot := NewBot(matrixClient, llmClient, testConfig(getTestStorageDir()))
	defer bot.Stop()

	evt := &event.Event{
		Type: event.EventMessage,
		Content: event.Content{
			Parsed: &event.MessageEventContent{
				MsgType: event.MsgVideo,
				URL:     "mxc://matrix.org/video789",
				Body:    "Video file",
			},
		},
	}

	_, _, err := bot.extractImageData(evt)
	if err == nil {
		t.Error("Expected error when message is video, not image")
	}
}

// TestExtractImageData_EmptyMXC verifies handling of empty MXC URI
func TestExtractImageData_EmptyMXC(t *testing.T) {
	defer setupTestEnv(t)()
	matrixClient, _ := matrix.NewClient("https://matrix.org", "@test:matrix.org", "test-token")
	llmClient := llm.NewClient("test-api-key", "claude-3-haiku-20240307", 100)
	bot := NewBot(matrixClient, llmClient, testConfig(getTestStorageDir()))
	defer bot.Stop()

	evt := &event.Event{
		Type: event.EventSticker,
		Content: event.Content{
			Parsed: &event.MessageEventContent{
				URL:  "", // Empty MXC URI
				Body: "Sticker without MXC",
			},
		},
	}

	mxcURI, body, err := bot.extractImageData(evt)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should still extract successfully (empty MXC will fail later in download)
	if mxcURI != "" {
		t.Errorf("Expected empty MXC URI, got %s", mxcURI)
	}

	if body != "Sticker without MXC" {
		t.Errorf("Expected body 'Sticker without MXC', got %s", body)
	}
}
