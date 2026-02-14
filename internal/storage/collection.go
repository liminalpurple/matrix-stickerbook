// Package storage provides persistent storage for collected stickers and curated packs.
// It manages two JSON files: collection.json (all collected stickers) and packs.json (pack definitions).
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AddSticker adds a new sticker to the collection
func AddSticker(dataDir string, sticker Sticker) error {
	collection, err := LoadCollection(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	// Check if sticker already exists (by ID)
	for i, existing := range collection.Stickers {
		if existing.ID == sticker.ID {
			// Update existing sticker
			collection.Stickers[i] = sticker
			return SaveCollection(dataDir, collection)
		}
	}

	// Add new sticker
	collection.Stickers = append(collection.Stickers, sticker)
	return SaveCollection(dataDir, collection)
}

// GetSticker retrieves a sticker by ID
func GetSticker(dataDir string, id string) (*Sticker, error) {
	collection, err := LoadCollection(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load collection: %w", err)
	}

	for _, sticker := range collection.Stickers {
		if sticker.ID == id {
			return &sticker, nil
		}
	}

	return nil, fmt.Errorf("sticker not found: %s", id)
}

// ListStickers returns all collected stickers
func ListStickers(dataDir string) ([]Sticker, error) {
	collection, err := LoadCollection(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load collection: %w", err)
	}

	return collection.Stickers, nil
}

// UpdateAltText updates the generated alt-text for a sticker
func UpdateAltText(dataDir string, id string, altText string) error {
	collection, err := LoadCollection(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	for i, sticker := range collection.Stickers {
		if sticker.ID == id {
			collection.Stickers[i].GeneratedAltText = altText
			return SaveCollection(dataDir, collection)
		}
	}

	return fmt.Errorf("sticker not found: %s", id)
}

// DeleteSticker removes a sticker from the collection and all packs
func DeleteSticker(dataDir string, id string) error {
	// Load collection
	collection, err := LoadCollection(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	// Find and remove sticker
	found := false
	var packNames []string
	for i, sticker := range collection.Stickers {
		if sticker.ID == id {
			// Remember which packs it's in
			packNames = sticker.InPacks
			// Remove from collection
			collection.Stickers = append(collection.Stickers[:i], collection.Stickers[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("sticker not found: %s", id)
	}

	// Save updated collection
	if err := SaveCollection(dataDir, collection); err != nil {
		return fmt.Errorf("failed to save collection: %w", err)
	}

	// Remove from all packs it was in
	for _, packName := range packNames {
		if err := RemoveFromPack(dataDir, packName, []string{id}); err != nil {
			// Log but don't fail - the sticker is already deleted from collection
			fmt.Printf("Warning: failed to remove sticker from pack %s: %v\n", packName, err)
		}
	}

	return nil
}

// LoadCollection loads the collection from disk
func LoadCollection(dataDir string) (*Collection, error) {
	collectionPath := filepath.Join(dataDir, "collection.json")

	// Check if file exists
	if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
		// Return empty collection if file doesn't exist
		return &Collection{Stickers: []Sticker{}}, nil
	}

	data, err := os.ReadFile(collectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read collection file: %w", err)
	}

	var collection Collection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("failed to unmarshal collection: %w", err)
	}

	return &collection, nil
}

// SaveCollection saves the collection to disk
func SaveCollection(dataDir string, collection *Collection) error {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	collectionPath := filepath.Join(dataDir, "collection.json")

	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal collection: %w", err)
	}

	if err := os.WriteFile(collectionPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write collection file: %w", err)
	}

	return nil
}

// SetStickerUsage sets the usage types for a specific sticker
func SetStickerUsage(dataDir string, stickerID string, usage []string) error {
	collection, err := LoadCollection(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	// Find the sticker
	for i, sticker := range collection.Stickers {
		if sticker.ID == stickerID {
			collection.Stickers[i].Usage = usage
			return SaveCollection(dataDir, collection)
		}
	}

	return fmt.Errorf("sticker not found: %s", stickerID)
}

// SetStickerName sets the shortcode name for a specific sticker
func SetStickerName(dataDir string, stickerID string, name string) error {
	collection, err := LoadCollection(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	// Find the sticker
	for i, sticker := range collection.Stickers {
		if sticker.ID == stickerID {
			collection.Stickers[i].Name = name
			return SaveCollection(dataDir, collection)
		}
	}

	return fmt.Errorf("sticker not found: %s", stickerID)
}
