package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/liminalpurple/matrix-stickerbook/internal/config"
	"github.com/liminalpurple/matrix-stickerbook/internal/llm"
	"github.com/liminalpurple/matrix-stickerbook/internal/matrix"
	"github.com/liminalpurple/matrix-stickerbook/internal/storage"
	"github.com/spf13/cobra"
)

// NewTestCmd creates the test command
func NewTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test Matrix connection and functionality",
		Long: `Test that all components are working correctly:

  - Configuration loads properly
  - Matrix connection and authentication
  - Media download/upload
  - Claude vision API (alt-text generation)
  - Storage operations

This is useful for verifying setup before running the bot.`,
		RunE: runTest,
	}
}

func runTest(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	fmt.Println("🧪 Running stickerbook tests...")
	fmt.Println()

	// Test 1: Load configuration
	fmt.Print("📋 Loading configuration... ")
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌\n   Error: %v\n", err)
		return err
	}
	fmt.Println("✅")

	// Verify required config
	if cfg.Matrix.AccessToken == "" {
		return fmt.Errorf("no access token configured")
	}
	if cfg.Anthropic.APIKey == "" {
		return fmt.Errorf("no Anthropic API key configured")
	}

	// Test 2: Matrix connection
	fmt.Print("🔌 Connecting to Matrix... ")
	matrixClient, err := matrix.NewClient(
		cfg.Matrix.Homeserver,
		cfg.Matrix.UserID,
		cfg.Matrix.AccessToken,
	)
	if err != nil {
		fmt.Printf("❌\n   Error: %v\n", err)
		return err
	}
	fmt.Println("✅")

	// Test 3: Verify credentials
	fmt.Print("🔑 Verifying credentials... ")
	if err := matrixClient.Connect(ctx); err != nil {
		fmt.Printf("❌\n   Error: %v\n", err)
		return err
	}
	fmt.Printf("✅\n   Logged in as: %s\n", matrixClient.UserID)
	fmt.Println()

	// Test 4: Create LLM client
	fmt.Print("🤖 Creating LLM client... ")
	llmClient := llm.NewClient(
		cfg.Anthropic.APIKey,
		cfg.Anthropic.Model,
		cfg.Anthropic.MaxTokens,
	)
	fmt.Printf("✅\n   Model: %s (max tokens: %d)\n", llmClient.Model(), llmClient.MaxTokens())
	fmt.Println()

	// Test 5: Generate test image and upload
	fmt.Print("🖼️  Creating test image... ")
	testImageData := createTestImage()
	fmt.Println("✅")

	fmt.Print("📤 Uploading test image to Matrix... ")
	testMXC, err := matrixClient.UploadMedia(ctx, testImageData, "image/png")
	if err != nil {
		fmt.Printf("❌\n   Error: %v\n", err)
		return err
	}
	fmt.Printf("✅\n   MXC URI: %s\n", testMXC)

	// Test 6: Download the image we just uploaded
	fmt.Print("📥 Downloading test image... ")
	downloadedData, downloadedMime, err := matrixClient.DownloadMedia(ctx, testMXC)
	if err != nil {
		fmt.Printf("❌\n   Error: %v\n", err)
		return err
	}
	fmt.Printf("✅\n   Size: %d bytes, MIME: %s\n", len(downloadedData), downloadedMime)

	// Test 7: Get image info
	fmt.Print("ℹ️  Extracting image info... ")
	imageInfo, err := matrix.GetImageInfo(downloadedData)
	if err != nil {
		fmt.Printf("❌\n   Error: %v\n", err)
		return err
	}
	fmt.Printf("✅\n   Dimensions: %dx%d, MIME: %s\n", imageInfo.Width, imageInfo.Height, imageInfo.MimeType)

	// Test 8: Generate alt-text
	fmt.Print("✨ Generating alt-text with Claude... ")
	altText, err := llmClient.GenerateAltText(ctx, downloadedData, imageInfo.MimeType)
	if err != nil {
		fmt.Printf("❌\n   Error: %v\n", err)
		return err
	}
	fmt.Printf("✅\n   Alt-text: %s\n", altText)
	fmt.Println()

	// Test 9: Storage operations
	fmt.Print("💾 Testing storage operations... ")

	// Create test sticker
	testSticker := storage.Sticker{
		ID:               matrix.HashImage(downloadedData),
		CollectedAt:      time.Now(),
		SourceRoom:       "!test:matrix.org",
		SourceEvent:      "$test-event",
		SourceMXC:        testMXC,
		LocalMXC:         testMXC,
		MimeType:         imageInfo.MimeType,
		Width:            imageInfo.Width,
		Height:           imageInfo.Height,
		SizeBytes:        imageInfo.SizeBytes,
		OriginalBody:     "Test sticker",
		GeneratedAltText: altText,
		InPacks:          []string{},
	}

	// Save to collection
	if err := storage.AddSticker(cfg.Storage.DataDir, testSticker); err != nil {
		fmt.Printf("❌\n   Error: %v\n", err)
		return err
	}

	// Retrieve it back
	retrieved, err := storage.GetSticker(cfg.Storage.DataDir, testSticker.ID)
	if err != nil {
		fmt.Printf("❌\n   Error: %v\n", err)
		return err
	}

	if retrieved.ID != testSticker.ID {
		fmt.Printf("❌\n   Error: Retrieved sticker ID mismatch\n")
		return fmt.Errorf("storage verification failed")
	}

	fmt.Printf("✅\n   Saved and retrieved test sticker: %s\n", testSticker.ID[:16]+"...")
	fmt.Println()

	// All tests passed!
	fmt.Println("🎉 All tests passed! The bot is ready to run.")
	fmt.Println()
	fmt.Println("To start the bot, run:")
	fmt.Println("  ./stickerbook bot")
	fmt.Println()

	return nil
}

// createTestImage generates a simple test PNG image (1x1 white pixel)
func createTestImage() []byte {
	// Minimal valid PNG: 1x1 white pixel
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1 dimensions
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41, // IDAT chunk
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xFF, 0xFF, 0x3F,
		0x00, 0x05, 0xFE, 0x02, 0xFE, 0xDC, 0xCC, 0x59,
		0xE7, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, // IEND chunk
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}
}
