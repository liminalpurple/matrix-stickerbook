package matrix

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/gif"  // Import for image format support
	_ "image/jpeg" // Import for image format support
	_ "image/png"  // Import for image format support

	"maunium.net/go/mautrix/id"
)

// ImageInfo contains metadata about an image
type ImageInfo struct {
	Width     int
	Height    int
	SizeBytes int64
	MimeType  string
}

// DownloadMedia downloads media from an MXC URI
func (c *Client) DownloadMedia(ctx context.Context, mxcURI string) ([]byte, string, error) {
	parsedURI, err := id.ParseContentURI(mxcURI)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse MXC URI: %w", err)
	}

	data, err := c.DownloadBytes(ctx, parsedURI)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download media: %w", err)
	}

	// Detect MIME type from data
	mimeType := detectMimeType(data)

	return data, mimeType, nil
}

// UploadMedia uploads media to the homeserver and returns the new MXC URI
func (c *Client) UploadMedia(ctx context.Context, data []byte, mimeType string) (string, error) {
	uploadResp, err := c.UploadBytes(ctx, data, mimeType)
	if err != nil {
		return "", fmt.Errorf("failed to upload media: %w", err)
	}

	return uploadResp.ContentURI.String(), nil
}

// GetImageInfo extracts image metadata
func GetImageInfo(data []byte) (*ImageInfo, error) {
	img, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	mimeType := formatToMimeType(format)

	return &ImageInfo{
		Width:     img.Width,
		Height:    img.Height,
		SizeBytes: int64(len(data)),
		MimeType:  mimeType,
	}, nil
}

// HashImage generates a SHA256 hash of image data (for sticker ID)
func HashImage(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// detectMimeType attempts to detect MIME type from data
func detectMimeType(data []byte) string {
	if len(data) < 4 {
		return "application/octet-stream"
	}

	// Check PNG signature
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png"
	}

	// Check JPEG signature
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg"
	}

	// Check GIF signature
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return "image/gif"
	}

	// Check WebP signature
	if len(data) >= 12 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
		data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
		return "image/webp"
	}

	return "application/octet-stream"
}

// formatToMimeType converts image format string to MIME type
func formatToMimeType(format string) string {
	switch format {
	case "png":
		return "image/png"
	case "jpeg":
		return "image/jpeg"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	default:
		return "image/" + format
	}
}
