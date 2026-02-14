package llm

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"testing"
)

// TestNewClient verifies client creation
func TestNewClient(t *testing.T) {
	client := NewClient("test-api-key", "claude-3-haiku-20240307", 100)

	if client == nil {
		t.Fatal("Expected client to be created")
	}

	if client.Model() != "claude-3-haiku-20240307" {
		t.Errorf("Expected model claude-3-haiku-20240307, got %s", client.Model())
	}

	if client.MaxTokens() != 100 {
		t.Errorf("Expected max tokens 100, got %d", client.MaxTokens())
	}
}

// TestIsImageMimeType_Valid verifies valid image MIME types
func TestIsImageMimeType_Valid(t *testing.T) {
	validTypes := []string{
		"image/png",
		"image/jpeg",
		"image/gif",
		"image/webp",
	}

	for _, mimeType := range validTypes {
		t.Run(mimeType, func(t *testing.T) {
			if !isImageMimeType(mimeType) {
				t.Errorf("Expected %s to be valid image MIME type", mimeType)
			}
		})
	}
}

// TestIsImageMimeType_Invalid verifies invalid MIME types are rejected
func TestIsImageMimeType_Invalid(t *testing.T) {
	invalidTypes := []string{
		"application/pdf",
		"text/plain",
		"video/mp4",
		"image/svg+xml",
		"",
	}

	for _, mimeType := range invalidTypes {
		t.Run(mimeType, func(t *testing.T) {
			if isImageMimeType(mimeType) {
				t.Errorf("Expected %s to be invalid image MIME type", mimeType)
			}
		})
	}
}

// TestGenerateAltText_EmptyData verifies error on empty image data
func TestGenerateAltText_EmptyData(t *testing.T) {
	client := NewClient("test-api-key", "claude-3-haiku-20240307", 100)

	_, err := client.GenerateAltText(context.Background(), []byte{}, "image/png")
	if err == nil {
		t.Error("Expected error when generating alt-text for empty data")
	}
}

// TestGenerateAltText_InvalidMimeType verifies error on invalid MIME type
func TestGenerateAltText_InvalidMimeType(t *testing.T) {
	client := NewClient("test-api-key", "claude-3-haiku-20240307", 100)

	// Create test image data
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("Failed to encode test image: %v", err)
	}

	_, err := client.GenerateAltText(context.Background(), buf.Bytes(), "application/pdf")
	if err == nil {
		t.Error("Expected error when generating alt-text for invalid MIME type")
	}
}

// Note: We don't test actual API calls here since that would require:
// 1. Real API credentials
// 2. Network access
// 3. API costs
//
// Integration tests with real API should be run separately with:
// - Valid API key in config
// - Network connectivity
// - Acceptance of API costs
//
// For unit testing, we verify:
// - Client creation
// - MIME type validation
// - Error handling for invalid inputs
// - Request structure (indirectly through error cases)
