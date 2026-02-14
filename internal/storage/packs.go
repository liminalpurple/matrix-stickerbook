package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CreatePack creates a new empty pack
func CreatePack(dataDir string, name string, displayName string) error {
	return CreatePackWithAttribution(dataDir, name, displayName, "")
}

// CreatePackWithAttribution creates a new empty pack with author attribution
func CreatePackWithAttribution(dataDir string, name string, displayName string, attribution string) error {
	packsData, err := LoadPacks(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load packs: %w", err)
	}

	// Check if pack already exists
	for _, pack := range packsData.Packs {
		if pack.Name == name {
			return fmt.Errorf("pack already exists: %s", name)
		}
	}

	// Create new pack
	newPack := Pack{
		Name:           name,
		DisplayName:    displayName,
		Attribution:    attribution,
		StickerIDs:     []string{},
		PublishedRooms: make(map[string]string),
	}

	packsData.Packs = append(packsData.Packs, newPack)
	return SavePacks(dataDir, packsData)
}

// AddToPack adds stickers to a pack
func AddToPack(dataDir string, packName string, stickerIDs []string) error {
	packsData, err := LoadPacks(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load packs: %w", err)
	}

	collection, err := LoadCollection(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	// Find the pack
	packIndex := -1
	for i, pack := range packsData.Packs {
		if pack.Name == packName {
			packIndex = i
			break
		}
	}

	if packIndex == -1 {
		return fmt.Errorf("pack not found: %s", packName)
	}

	// Verify all stickers exist and add to pack
	for _, stickerID := range stickerIDs {
		// Check if sticker exists in collection
		found := false
		for _, sticker := range collection.Stickers {
			if sticker.ID == stickerID {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("sticker not found in collection: %s", stickerID)
		}

		// Check if sticker is already in pack
		alreadyInPack := false
		for _, existingID := range packsData.Packs[packIndex].StickerIDs {
			if existingID == stickerID {
				alreadyInPack = true
				break
			}
		}

		if !alreadyInPack {
			packsData.Packs[packIndex].StickerIDs = append(packsData.Packs[packIndex].StickerIDs, stickerID)
		}
	}

	// Update sticker's InPacks field
	for i := range collection.Stickers {
		for _, stickerID := range stickerIDs {
			if collection.Stickers[i].ID == stickerID {
				// Check if pack is already in sticker's InPacks
				inPacks := false
				for _, packInList := range collection.Stickers[i].InPacks {
					if packInList == packName {
						inPacks = true
						break
					}
				}
				if !inPacks {
					collection.Stickers[i].InPacks = append(collection.Stickers[i].InPacks, packName)
				}
			}
		}
	}

	if err := SaveCollection(dataDir, collection); err != nil {
		return fmt.Errorf("failed to update collection: %w", err)
	}

	return SavePacks(dataDir, packsData)
}

// RemoveFromPack removes stickers from a pack
func RemoveFromPack(dataDir string, packName string, stickerIDs []string) error {
	packsData, err := LoadPacks(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load packs: %w", err)
	}

	collection, err := LoadCollection(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	// Find the pack
	packIndex := -1
	for i, pack := range packsData.Packs {
		if pack.Name == packName {
			packIndex = i
			break
		}
	}

	if packIndex == -1 {
		return fmt.Errorf("pack not found: %s", packName)
	}

	// Remove stickers from pack
	for _, stickerID := range stickerIDs {
		newStickerIDs := []string{}
		for _, existingID := range packsData.Packs[packIndex].StickerIDs {
			if existingID != stickerID {
				newStickerIDs = append(newStickerIDs, existingID)
			}
		}
		packsData.Packs[packIndex].StickerIDs = newStickerIDs
	}

	// Update sticker's InPacks field
	for i := range collection.Stickers {
		for _, stickerID := range stickerIDs {
			if collection.Stickers[i].ID == stickerID {
				newInPacks := []string{}
				for _, packInList := range collection.Stickers[i].InPacks {
					if packInList != packName {
						newInPacks = append(newInPacks, packInList)
					}
				}
				collection.Stickers[i].InPacks = newInPacks
			}
		}
	}

	if err := SaveCollection(dataDir, collection); err != nil {
		return fmt.Errorf("failed to update collection: %w", err)
	}

	return SavePacks(dataDir, packsData)
}

// GetPack retrieves a pack by name
func GetPack(dataDir string, name string) (*Pack, error) {
	packsData, err := LoadPacks(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load packs: %w", err)
	}

	for _, pack := range packsData.Packs {
		if pack.Name == name {
			return &pack, nil
		}
	}

	return nil, fmt.Errorf("pack not found: %s", name)
}

// UpdatePublished records that a pack has been published to a room
func UpdatePublished(dataDir string, packName string, roomID string, stateKey string) error {
	packsData, err := LoadPacks(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load packs: %w", err)
	}

	// Find the pack
	for i, pack := range packsData.Packs {
		if pack.Name == packName {
			if packsData.Packs[i].PublishedRooms == nil {
				packsData.Packs[i].PublishedRooms = make(map[string]string)
			}
			packsData.Packs[i].PublishedRooms[roomID] = stateKey
			return SavePacks(dataDir, packsData)
		}
	}

	return fmt.Errorf("pack not found: %s", packName)
}

// SetPackAvatar sets the avatar URL for a pack
func SetPackAvatar(dataDir string, packName string, avatarURL string) error {
	packsData, err := LoadPacks(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load packs: %w", err)
	}

	// Find the pack
	for i, pack := range packsData.Packs {
		if pack.Name == packName {
			packsData.Packs[i].AvatarURL = avatarURL
			return SavePacks(dataDir, packsData)
		}
	}

	return fmt.Errorf("pack not found: %s", packName)
}

// LoadPacks loads pack definitions from disk
func LoadPacks(dataDir string) (*PacksData, error) {
	packsPath := filepath.Join(dataDir, "packs.json")

	// Check if file exists
	if _, err := os.Stat(packsPath); os.IsNotExist(err) {
		// Return empty packs data if file doesn't exist
		return &PacksData{Packs: []Pack{}}, nil
	}

	data, err := os.ReadFile(packsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read packs file: %w", err)
	}

	var packsData PacksData
	if err := json.Unmarshal(data, &packsData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal packs: %w", err)
	}

	return &packsData, nil
}

// SavePacks saves pack definitions to disk
func SavePacks(dataDir string, packsData *PacksData) error {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	packsPath := filepath.Join(dataDir, "packs.json")

	data, err := json.MarshalIndent(packsData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal packs: %w", err)
	}

	if err := os.WriteFile(packsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write packs file: %w", err)
	}

	return nil
}

// ListPacks returns all packs
func ListPacks(dataDir string) ([]Pack, error) {
	packsData, err := LoadPacks(dataDir)
	if err != nil {
		return nil, err
	}

	return packsData.Packs, nil
}

// SetPackUsage sets the default usage for all stickers in a pack
func SetPackUsage(dataDir string, packName string, usage []string) error {
	packsData, err := LoadPacks(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load packs: %w", err)
	}

	// Find the pack
	for i, pack := range packsData.Packs {
		if pack.Name == packName {
			packsData.Packs[i].Usage = usage
			return SavePacks(dataDir, packsData)
		}
	}

	return fmt.Errorf("pack not found: %s", packName)
}
