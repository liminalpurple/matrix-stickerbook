package storage

import "time"

// Sticker represents a collected sticker
type Sticker struct {
	ID               string    `json:"id"`                 // SHA256 hash of image data (internal ID)
	Name             string    `json:"name"`               // Shortcode name for emoji (defaults to ID)
	CollectedAt      time.Time `json:"collected_at"`       // When sticker was collected
	SourceRoom       string    `json:"source_room"`        // Room ID where found
	SourceEvent      string    `json:"source_event"`       // Event ID of original message
	SourceMXC        string    `json:"source_mxc"`         // Original MXC URI
	LocalMXC         string    `json:"local_mxc"`          // Rehosted MXC URI
	MimeType         string    `json:"mime_type"`          // Image MIME type
	Width            int       `json:"width"`              // Image width in pixels
	Height           int       `json:"height"`             // Image height in pixels
	SizeBytes        int64     `json:"size_bytes"`         // File size in bytes
	OriginalBody     string    `json:"original_body"`      // Original description/alt-text
	GeneratedAltText string    `json:"generated_alt_text"` // Claude-generated alt-text
	InPacks          []string  `json:"in_packs"`           // Pack names containing this sticker
	Usage            []string  `json:"usage,omitempty"`    // Usage types: "sticker", "emoticon", or both
}

// Collection holds all collected stickers
type Collection struct {
	Stickers []Sticker `json:"stickers"`
}

// Pack represents a curated sticker pack
type Pack struct {
	Name           string            `json:"name"`                      // Internal pack name
	DisplayName    string            `json:"display_name"`              // User-facing display name
	AvatarURL      string            `json:"avatar_url,omitempty"`      // Pack icon MXC URI
	Attribution    string            `json:"attribution,omitempty"`     // Pack author (Matrix ID)
	StickerIDs     []string          `json:"sticker_ids"`               // Sticker IDs in this pack
	PublishedRooms map[string]string `json:"published_rooms,omitempty"` // Room ID -> state key mapping
	Usage          []string          `json:"usage,omitempty"`           // Default usage for pack: "sticker", "emoticon", or both
}

// PacksData holds all pack definitions
type PacksData struct {
	Packs []Pack `json:"packs"`
}
