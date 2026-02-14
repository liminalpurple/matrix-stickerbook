package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/liminalpurple/matrix-stickerbook/internal/matrix"
	"github.com/liminalpurple/matrix-stickerbook/internal/storage"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// validCommands are the reaction commands we respond to
var validCommands = map[string]bool{
	"!yoink": true,
	"!nom":   true,
	"!grab":  true,
}

// processReaction handles a reaction event and collects the sticker if appropriate
func (b *Bot) processReaction(ctx context.Context, evt *event.Event) error {
	// Parse reaction content
	content, ok := evt.Content.Parsed.(*event.ReactionEventContent)
	if !ok {
		return fmt.Errorf("failed to parse reaction content")
	}

	// Check if this is one of our collection commands
	reaction := content.RelatesTo.Key
	if !validCommands[reaction] {
		return nil // Not a collection command, ignore
	}

	log.Printf("Detected %s command from %s in room %s", reaction, evt.Sender, evt.RoomID)

	// Get the parent event that was reacted to
	parentEventID := content.RelatesTo.EventID
	parentEvent, err := b.client.GetEvent(ctx, evt.RoomID, parentEventID)
	if err != nil {
		return fmt.Errorf("failed to get parent event: %w", err)
	}

	// Extract image data from parent event
	mxcURI, body, err := b.extractImageData(parentEvent)
	if err != nil {
		return fmt.Errorf("parent event is not a valid image/sticker: %w", err)
	}

	log.Printf("Collecting sticker: %s (MXC: %s)", body, mxcURI)

	// Run the collection workflow
	if err := b.collectSticker(ctx, evt.RoomID, parentEventID, mxcURI, body); err != nil {
		return fmt.Errorf("failed to collect sticker: %w", err)
	}

	// Redact the reaction to confirm collection (cleaner timeline)
	if err := b.redactReaction(ctx, evt.RoomID, evt.ID); err != nil {
		log.Printf("Warning: failed to redact reaction: %v", err)
	}

	return nil
}

// extractImageData extracts the MXC URI and body text from an image or sticker event
func (b *Bot) extractImageData(evt *event.Event) (mxcURI id.ContentURIString, body string, err error) {
	// Handle both m.sticker and m.room.message (with msgtype=m.image)
	switch evt.Type {
	case event.EventSticker:
		// m.sticker events might not be parsed - try parsed first, fall back to raw
		if content, ok := evt.Content.Parsed.(*event.MessageEventContent); ok {
			return content.URL, content.Body, nil
		}

		// Fall back to raw content access
		url, ok := evt.Content.Raw["url"].(string)
		if !ok {
			return "", "", fmt.Errorf("sticker missing url field")
		}

		body, _ := evt.Content.Raw["body"].(string)
		return id.ContentURIString(url), body, nil

	case event.EventMessage:
		// Try parsed content first
		if content, ok := evt.Content.Parsed.(*event.MessageEventContent); ok {
			if content.MsgType != event.MsgImage {
				return "", "", fmt.Errorf("message is not an image (msgtype=%s)", content.MsgType)
			}
			return content.URL, content.Body, nil
		}

		// Fall back to raw content access
		msgtype, ok := evt.Content.Raw["msgtype"].(string)
		if !ok || msgtype != "m.image" {
			return "", "", fmt.Errorf("message is not an image (msgtype=%s)", msgtype)
		}

		url, ok := evt.Content.Raw["url"].(string)
		if !ok {
			return "", "", fmt.Errorf("message missing url field")
		}

		body, _ := evt.Content.Raw["body"].(string)
		return id.ContentURIString(url), body, nil

	default:
		return "", "", fmt.Errorf("unsupported event type: %s", evt.Type.Type)
	}
}

// collectSticker downloads, rehosts, generates alt-text, and saves a sticker
func (b *Bot) collectSticker(ctx context.Context, roomID id.RoomID, eventID id.EventID, mxcURI id.ContentURIString, originalBody string) error {
	// Check if media is already on our homeserver
	parsedMXC, err := mxcURI.Parse()
	if err != nil {
		return fmt.Errorf("invalid MXC URI: %w", err)
	}

	localMXC := string(mxcURI)
	needsRehost := parsedMXC.Homeserver != b.client.UserID.Homeserver()

	// Download image from source MXC URI
	imageData, detectedMimeType, err := b.client.DownloadMedia(ctx, string(mxcURI))
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Get image info (dimensions, MIME type, size)
	imageInfo, err := matrix.GetImageInfo(imageData)
	if err != nil {
		return fmt.Errorf("failed to get image info: %w", err)
	}

	// Use detected MIME type from download if GetImageInfo didn't detect it properly
	if imageInfo.MimeType == "" || imageInfo.MimeType == "application/octet-stream" {
		imageInfo.MimeType = detectedMimeType
	}

	// Generate sticker ID from hash
	stickerID := matrix.HashImage(imageData)

	log.Printf("Image info: %dx%d, %s, %d bytes, ID=%s",
		imageInfo.Width, imageInfo.Height, imageInfo.MimeType, imageInfo.SizeBytes, stickerID)

	// Upload to local homeserver if needed (rehost)
	if needsRehost {
		localMXC, err = b.client.UploadMedia(ctx, imageData, imageInfo.MimeType)
		if err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}
		log.Printf("Rehosted: %s → %s", mxcURI, localMXC)
	} else {
		log.Printf("Already on local homeserver: %s", mxcURI)
	}

	// Generate alt-text using Claude
	altText, err := b.llmClient.GenerateAltText(ctx, imageData, imageInfo.MimeType)
	if err != nil {
		return fmt.Errorf("alt-text generation failed: %w", err)
	}

	// Clean up alt-text: replace linebreaks with spaces and trim
	altText = strings.ReplaceAll(altText, "\r\n", " ")
	altText = strings.ReplaceAll(altText, "\n", " ")
	altText = strings.ReplaceAll(altText, "\r", " ")
	altText = strings.TrimSpace(altText)

	log.Printf("Generated alt-text: %s", altText)

	// Create sticker record
	sticker := storage.Sticker{
		ID:               stickerID,
		Name:             stickerID, // Default to SHA256 hash
		CollectedAt:      time.Now(),
		SourceRoom:       roomID.String(),
		SourceEvent:      eventID.String(),
		SourceMXC:        string(mxcURI),
		LocalMXC:         localMXC,
		MimeType:         imageInfo.MimeType,
		Width:            imageInfo.Width,
		Height:           imageInfo.Height,
		SizeBytes:        imageInfo.SizeBytes,
		OriginalBody:     originalBody,
		GeneratedAltText: altText,
		InPacks:          []string{},
	}

	// Save to collection
	if err := storage.AddSticker(b.storageDir, sticker); err != nil {
		return fmt.Errorf("failed to save sticker: %w", err)
	}

	log.Printf("✅ Sticker collected successfully: %s", stickerID)

	return nil
}

// redactReaction redacts the reaction event to confirm collection
func (b *Bot) redactReaction(ctx context.Context, roomID id.RoomID, reactionEventID id.EventID) error {
	_, err := b.client.RedactEvent(ctx, roomID, reactionEventID)
	return err
}
