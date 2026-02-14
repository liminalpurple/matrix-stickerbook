package llm

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)

const defaultPrompt = `Describe this sticker in one short sentence.
Aim for ~15 words, max 30 words unless the image contains text.
Focus on: main subject, emotion/action, distinctive shapes/colors, clothing/art style.
IMPORTANT: If there is any text visible in the image, include it verbatim (for accessibility).
Output ONLY the description - no markdown, no headers, no formatting.

Good examples:
"Anime girl with cat ears and school uniform looking surprised"
"Two characters in spacesuits kissing against starry background"
"Bright pink octopus wearing top hat with text 'Nope' in bold letters"`

// GenerateAltText generates alt-text description for an image using Claude vision
func (c *Client) GenerateAltText(ctx context.Context, imageData []byte, mimeType string) (string, error) {
	if len(imageData) == 0 {
		return "", fmt.Errorf("image data is empty")
	}

	// Validate MIME type is an image
	if !isImageMimeType(mimeType) {
		return "", fmt.Errorf("invalid MIME type for image: %s", mimeType)
	}

	// Encode image to base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	// Create vision request
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: c.maxTokens,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewImageBlockBase64(mimeType, base64Image),
				anthropic.NewTextBlock(defaultPrompt),
			),
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate alt-text: %w", err)
	}

	// Extract text from response
	if len(message.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	// The response should contain a text block - the union has all fields
	if message.Content[0].Type != "text" {
		return "", fmt.Errorf("unexpected response type: %s", message.Content[0].Type)
	}

	return message.Content[0].Text, nil
}

// isImageMimeType checks if the MIME type is a valid image type
func isImageMimeType(mimeType string) bool {
	validTypes := []string{
		"image/png",
		"image/jpeg",
		"image/gif",
		"image/webp",
	}

	for _, valid := range validTypes {
		if mimeType == valid {
			return true
		}
	}
	return false
}
