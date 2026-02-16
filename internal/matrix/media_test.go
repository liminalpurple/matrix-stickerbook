package matrix

import (
	"bytes"
	"image"
	"image/png"
	"testing"
)

// TestHashImage_Consistency verifies same data produces same hash
func TestHashImage_Consistency(t *testing.T) {
	data := []byte("test data for hashing")
	hash1 := HashImage(data)
	hash2 := HashImage(data)

	if hash1 != hash2 {
		t.Error("Same data produced different hashes")
	}
}

// TestHashImage_Uniqueness verifies different data produces different hashes
func TestHashImage_Uniqueness(t *testing.T) {
	data1 := []byte("first dataset")
	data2 := []byte("second dataset")

	hash1 := HashImage(data1)
	hash2 := HashImage(data2)

	if hash1 == hash2 {
		t.Error("Different data produced same hash")
	}
}

// TestHashImage_Format verifies hash has correct format
func TestHashImage_Format(t *testing.T) {
	data := []byte("test")
	hash := HashImage(data)

	// Should be 64 hex characters (SHA256 hash)
	if len(hash) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash))
	}

	// Verify it's valid hex
	for _, c := range hash {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			t.Errorf("Hash contains invalid hex character: %c", c)
			break
		}
	}
}

// TestHashImage_EmptyData verifies hashing empty data works
func TestHashImage_EmptyData(t *testing.T) {
	data := []byte{}
	hash := HashImage(data)

	if len(hash) != 64 {
		t.Errorf("Expected hash length 64 for empty data, got %d", len(hash))
	}

	// Verify it's valid hex
	for _, c := range hash {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			t.Errorf("Hash contains invalid hex character: %c", c)
			break
		}
	}
}

// TestGetImageInfo_ValidPNG verifies extracting info from valid PNG
func TestGetImageInfo_ValidPNG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 512, 512))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("Failed to encode test image: %v", err)
	}
	data := buf.Bytes()

	info, err := GetImageInfo(data)
	if err != nil {
		t.Fatalf("Failed to get image info: %v", err)
	}

	if info.Width != 512 {
		t.Errorf("Expected width 512, got %d", info.Width)
	}
	if info.Height != 512 {
		t.Errorf("Expected height 512, got %d", info.Height)
	}
	if info.MimeType != "image/png" {
		t.Errorf("Expected MIME type image/png, got %s", info.MimeType)
	}
	if info.SizeBytes != int64(len(data)) {
		t.Errorf("Expected size %d, got %d", len(data), info.SizeBytes)
	}
}

// TestGetImageInfo_DifferentDimensions verifies dimension detection works
func TestGetImageInfo_DifferentDimensions(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"square", 256, 256},
		{"landscape", 1024, 768},
		{"portrait", 600, 800},
		{"small", 64, 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, tt.width, tt.height))
			var buf bytes.Buffer
			if err := png.Encode(&buf, img); err != nil {
				t.Fatalf("Failed to encode image: %v", err)
			}

			info, err := GetImageInfo(buf.Bytes())
			if err != nil {
				t.Fatalf("Failed to get image info: %v", err)
			}

			if info.Width != tt.width {
				t.Errorf("Expected width %d, got %d", tt.width, info.Width)
			}
			if info.Height != tt.height {
				t.Errorf("Expected height %d, got %d", tt.height, info.Height)
			}
		})
	}
}

// TestGetImageInfo_InvalidData verifies error on corrupted image
func TestGetImageInfo_InvalidData(t *testing.T) {
	invalidData := []byte("this is not a valid image")

	_, err := GetImageInfo(invalidData)
	if err == nil {
		t.Error("Expected error when getting info from invalid image data")
	}
}

// TestGetImageInfo_EmptyData verifies error on empty data
func TestGetImageInfo_EmptyData(t *testing.T) {
	_, err := GetImageInfo([]byte{})
	if err == nil {
		t.Error("Expected error when getting info from empty data")
	}
}

// TestDetectMimeType_PNG verifies PNG detection
func TestDetectMimeType_PNG(t *testing.T) {
	pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	mimeType := detectMimeType(pngSignature)

	if mimeType != "image/png" {
		t.Errorf("Expected image/png, got %s", mimeType)
	}
}

// TestDetectMimeType_JPEG verifies JPEG detection
func TestDetectMimeType_JPEG(t *testing.T) {
	jpegSignature := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
	mimeType := detectMimeType(jpegSignature)

	if mimeType != "image/jpeg" {
		t.Errorf("Expected image/jpeg, got %s", mimeType)
	}
}

// TestDetectMimeType_GIF verifies GIF detection
func TestDetectMimeType_GIF(t *testing.T) {
	gifSignature := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}
	mimeType := detectMimeType(gifSignature)

	if mimeType != "image/gif" {
		t.Errorf("Expected image/gif, got %s", mimeType)
	}
}

// TestDetectMimeType_WebP verifies WebP detection
func TestDetectMimeType_WebP(t *testing.T) {
	webpSignature := []byte{
		0x52, 0x49, 0x46, 0x46, // "RIFF"
		0x00, 0x00, 0x00, 0x00, // size (doesn't matter for test)
		0x57, 0x45, 0x42, 0x50, // "WEBP"
	}
	mimeType := detectMimeType(webpSignature)

	if mimeType != "image/webp" {
		t.Errorf("Expected image/webp, got %s", mimeType)
	}
}

// TestDetectMimeType_Unknown verifies fallback for unknown formats
func TestDetectMimeType_Unknown(t *testing.T) {
	unknownData := []byte{0x00, 0x00, 0x00, 0x00}
	mimeType := detectMimeType(unknownData)

	if mimeType != "application/octet-stream" {
		t.Errorf("Expected application/octet-stream for unknown data, got %s", mimeType)
	}
}

// TestDetectMimeType_TooShort verifies fallback for insufficient data
func TestDetectMimeType_TooShort(t *testing.T) {
	shortData := []byte{0x89, 0x50}
	mimeType := detectMimeType(shortData)

	if mimeType != "application/octet-stream" {
		t.Errorf("Expected application/octet-stream for short data, got %s", mimeType)
	}
}

// TestDetectMimeType_Empty verifies fallback for empty data
func TestDetectMimeType_Empty(t *testing.T) {
	mimeType := detectMimeType([]byte{})

	if mimeType != "application/octet-stream" {
		t.Errorf("Expected application/octet-stream for empty data, got %s", mimeType)
	}
}

// TestFormatToMimeType_Common verifies common format conversions
func TestFormatToMimeType_Common(t *testing.T) {
	tests := []struct {
		format   string
		expected string
	}{
		{"png", "image/png"},
		{"jpeg", "image/jpeg"},
		{"jpg", "image/jpg"},
		{"gif", "image/gif"},
		{"webp", "image/webp"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := formatToMimeType(tt.format)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestFormatToMimeType_Unknown verifies unknown format handling
func TestFormatToMimeType_Unknown(t *testing.T) {
	result := formatToMimeType("unknown")
	expected := "image/unknown"

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestFormatToMimeType_Empty verifies empty format handling
func TestFormatToMimeType_Empty(t *testing.T) {
	result := formatToMimeType("")
	expected := "image/"

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
