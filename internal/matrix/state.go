package matrix

import (
	"context"
	"fmt"

	"github.com/liminalpurple/matrix-stickerbook/internal/storage"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// MSC2545 Sticker Pack Types

// PackInfo represents the pack metadata
type PackInfo struct {
	DisplayName string   `json:"display_name"`
	AvatarURL   string   `json:"avatar_url,omitempty"`
	Usage       []string `json:"usage,omitempty"`
	Attribution string   `json:"attribution,omitempty"`
}

// StickerData represents a single sticker in the pack
type StickerData struct {
	URL   string   `json:"url"`
	Body  string   `json:"body"`
	Usage []string `json:"usage,omitempty"` // Per-sticker usage override
	Info  struct {
		Width    int    `json:"w"`
		Height   int    `json:"h"`
		Size     int64  `json:"size"`
		MimeType string `json:"mimetype"`
	} `json:"info"`
}

// PackContent represents the MSC2545 state event content
type PackContent struct {
	Pack   PackInfo               `json:"pack"`
	Images map[string]StickerData `json:"images"`
}

// PublishPack publishes a sticker pack to a Matrix room as an MSC2545 state event
func (c *Client) PublishPack(ctx context.Context, dataDir string, packName string, roomID id.RoomID) error {
	// Load pack
	pack, err := storage.GetPack(dataDir, packName)
	if err != nil {
		return fmt.Errorf("failed to load pack: %w", err)
	}

	// Load collection to get sticker details
	collection, err := storage.LoadCollection(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	// Build images map
	images := make(map[string]StickerData)
	for _, stickerID := range pack.StickerIDs {
		// Find sticker in collection
		var sticker *storage.Sticker
		for i := range collection.Stickers {
			if collection.Stickers[i].ID == stickerID {
				sticker = &collection.Stickers[i]
				break
			}
		}

		if sticker == nil {
			return fmt.Errorf("sticker not found in collection: %s", stickerID)
		}

		// Use alt-text if available, otherwise original body
		body := sticker.GeneratedAltText
		if body == "" {
			body = sticker.OriginalBody
		}

		stickerData := StickerData{
			URL:  sticker.LocalMXC,
			Body: body,
		}
		stickerData.Info.Width = sticker.Width
		stickerData.Info.Height = sticker.Height
		stickerData.Info.Size = sticker.SizeBytes
		stickerData.Info.MimeType = sticker.MimeType

		// Include per-sticker usage if set (overrides pack default)
		if len(sticker.Usage) > 0 {
			stickerData.Usage = sticker.Usage
		}

		// Use Name as the shortcode key (defaults to SHA256 if not set)
		shortcode := sticker.Name
		if shortcode == "" {
			shortcode = stickerID
		}
		images[shortcode] = stickerData
	}

	// Build pack content
	packInfo := PackInfo{
		DisplayName: pack.DisplayName,
		Usage:       []string{"sticker", "emoticon"}, // Default to both unless pack.Usage is set
	}

	// Use pack's configured usage if set
	if len(pack.Usage) > 0 {
		packInfo.Usage = pack.Usage
	}

	// Add optional fields if present
	if pack.AvatarURL != "" {
		packInfo.AvatarURL = pack.AvatarURL
	}
	if pack.Attribution != "" {
		packInfo.Attribution = pack.Attribution
	}

	content := PackContent{
		Pack:   packInfo,
		Images: images,
	}

	// State key is the pack name
	stateKey := packName

	// Send state event
	_, err = c.SendStateEvent(ctx, roomID, event.Type{Type: "im.ponies.room_emotes", Class: event.StateEventType}, stateKey, content)
	if err != nil {
		return fmt.Errorf("failed to send state event: %w", err)
	}

	// Update pack's published rooms
	if err := storage.UpdatePublished(dataDir, packName, roomID.String(), stateKey); err != nil {
		return fmt.Errorf("failed to update published rooms: %w", err)
	}

	return nil
}
