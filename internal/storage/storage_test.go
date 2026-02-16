package storage

import (
	"os"
	"testing"
	"time"
)

// TestAddSticker_NewSticker verifies adding a new sticker works
func TestAddSticker_NewSticker(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	sticker := testSticker("sha256:abc123")
	if err := AddSticker(tmpDir, sticker); err != nil {
		t.Fatalf("Failed to add sticker: %v", err)
	}

	retrieved, err := GetSticker(tmpDir, "sha256:abc123")
	if err != nil {
		t.Fatalf("Failed to get sticker: %v", err)
	}

	if retrieved.ID != sticker.ID {
		t.Errorf("Expected ID %s, got %s", sticker.ID, retrieved.ID)
	}
	if retrieved.LocalMXC != sticker.LocalMXC {
		t.Errorf("Expected MXC %s, got %s", sticker.LocalMXC, retrieved.LocalMXC)
	}
}

// TestAddSticker_Duplicate verifies that adding the same sticker ID updates it
func TestAddSticker_Duplicate(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	sticker := testSticker("sha256:abc123")
	sticker.GeneratedAltText = "first version"

	if err := AddSticker(tmpDir, sticker); err != nil {
		t.Fatalf("Failed to add sticker: %v", err)
	}

	// Add again with different alt-text
	sticker.GeneratedAltText = "updated version"
	if err := AddSticker(tmpDir, sticker); err != nil {
		t.Fatalf("Failed to update sticker: %v", err)
	}

	// Should only have one sticker
	stickers, err := ListStickers(tmpDir)
	if err != nil {
		t.Fatalf("Failed to list stickers: %v", err)
	}

	if len(stickers) != 1 {
		t.Errorf("Expected 1 sticker after duplicate add, got %d", len(stickers))
	}

	if stickers[0].GeneratedAltText != "updated version" {
		t.Errorf("Expected updated alt-text, got %s", stickers[0].GeneratedAltText)
	}
}

// TestGetSticker_NotFound verifies error when sticker doesn't exist
func TestGetSticker_NotFound(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	_, err := GetSticker(tmpDir, "sha256:doesnotexist")
	if err == nil {
		t.Error("Expected error when getting non-existent sticker")
	}
}

// TestListStickers_Empty verifies empty collection returns empty list
func TestListStickers_Empty(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	stickers, err := ListStickers(tmpDir)
	if err != nil {
		t.Fatalf("Failed to list stickers: %v", err)
	}

	if len(stickers) != 0 {
		t.Errorf("Expected 0 stickers in empty collection, got %d", len(stickers))
	}
}

// TestUpdateAltText_NotFound verifies error when sticker doesn't exist
func TestUpdateAltText_NotFound(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	err := UpdateAltText(tmpDir, "sha256:doesnotexist", "new text")
	if err == nil {
		t.Error("Expected error when updating non-existent sticker")
	}
}

// TestUpdateAltText_Success verifies alt-text update works
func TestUpdateAltText_Success(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	sticker := testSticker("sha256:abc123")
	sticker.GeneratedAltText = "original"
	if err := AddSticker(tmpDir, sticker); err != nil {
		t.Fatalf("Failed to add sticker: %v", err)
	}

	if err := UpdateAltText(tmpDir, "sha256:abc123", "updated"); err != nil {
		t.Fatalf("Failed to update alt-text: %v", err)
	}

	retrieved, err := GetSticker(tmpDir, "sha256:abc123")
	if err != nil {
		t.Fatalf("Failed to get sticker: %v", err)
	}

	if retrieved.GeneratedAltText != "updated" {
		t.Errorf("Expected alt-text 'updated', got %s", retrieved.GeneratedAltText)
	}
}

// TestCreatePack_Success verifies pack creation
func TestCreatePack_Success(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if err := CreatePack(tmpDir, "favourites", "My Favourites"); err != nil {
		t.Fatalf("Failed to create pack: %v", err)
	}

	pack, err := GetPack(tmpDir, "favourites")
	if err != nil {
		t.Fatalf("Failed to get pack: %v", err)
	}

	if pack.Name != "favourites" {
		t.Errorf("Expected pack name 'favourites', got %s", pack.Name)
	}
	if pack.DisplayName != "My Favourites" {
		t.Errorf("Expected display name 'My Favourites', got %s", pack.DisplayName)
	}
	if len(pack.StickerIDs) != 0 {
		t.Errorf("Expected new pack to have 0 stickers, got %d", len(pack.StickerIDs))
	}
}

// TestCreatePack_Duplicate verifies error when pack already exists
func TestCreatePack_Duplicate(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if err := CreatePack(tmpDir, "favourites", "My Favourites"); err != nil {
		t.Fatalf("Failed to create pack: %v", err)
	}

	err := CreatePack(tmpDir, "favourites", "Different Name")
	if err == nil {
		t.Error("Expected error when creating duplicate pack")
	}
}

// TestGetPack_NotFound verifies error when pack doesn't exist
func TestGetPack_NotFound(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	_, err := GetPack(tmpDir, "doesnotexist")
	if err == nil {
		t.Error("Expected error when getting non-existent pack")
	}
}

// TestAddToPack_StickerNotInCollection verifies error when sticker doesn't exist
func TestAddToPack_StickerNotInCollection(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if err := CreatePack(tmpDir, "favourites", "My Favourites"); err != nil {
		t.Fatalf("Failed to create pack: %v", err)
	}

	err := AddToPack(tmpDir, "favourites", []string{"sha256:doesnotexist"})
	if err == nil {
		t.Error("Expected error when adding non-existent sticker to pack")
	}
}

// TestAddToPack_PackNotFound verifies error when pack doesn't exist
func TestAddToPack_PackNotFound(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	sticker := testSticker("sha256:abc123")
	if err := AddSticker(tmpDir, sticker); err != nil {
		t.Fatalf("Failed to add sticker: %v", err)
	}

	err := AddToPack(tmpDir, "doesnotexist", []string{"sha256:abc123"})
	if err == nil {
		t.Error("Expected error when adding to non-existent pack")
	}
}

// TestAddToPack_BidirectionalReferences verifies pack ↔ sticker references stay consistent
func TestAddToPack_BidirectionalReferences(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Add stickers
	sticker1 := testSticker("sha256:sticker1")
	sticker2 := testSticker("sha256:sticker2")
	if err := AddSticker(tmpDir, sticker1); err != nil {
		t.Fatalf("Failed to add sticker1: %v", err)
	}
	if err := AddSticker(tmpDir, sticker2); err != nil {
		t.Fatalf("Failed to add sticker2: %v", err)
	}

	// Create pack
	if err := CreatePack(tmpDir, "favourites", "My Favourites"); err != nil {
		t.Fatalf("Failed to create pack: %v", err)
	}

	// Add stickers to pack
	if err := AddToPack(tmpDir, "favourites", []string{"sha256:sticker1", "sha256:sticker2"}); err != nil {
		t.Fatalf("Failed to add stickers to pack: %v", err)
	}

	// Verify pack contains stickers
	pack, err := GetPack(tmpDir, "favourites")
	if err != nil {
		t.Fatalf("Failed to get pack: %v", err)
	}
	if len(pack.StickerIDs) != 2 {
		t.Errorf("Expected pack to contain 2 stickers, got %d", len(pack.StickerIDs))
	}

	// Verify stickers know they're in pack
	retrieved1, err := GetSticker(tmpDir, "sha256:sticker1")
	if err != nil {
		t.Fatalf("Failed to get sticker1: %v", err)
	}
	if !containsString(retrieved1.InPacks, "favourites") {
		t.Error("Expected sticker1 to be in 'favourites' pack")
	}

	retrieved2, err := GetSticker(tmpDir, "sha256:sticker2")
	if err != nil {
		t.Fatalf("Failed to get sticker2: %v", err)
	}
	if !containsString(retrieved2.InPacks, "favourites") {
		t.Error("Expected sticker2 to be in 'favourites' pack")
	}
}

// TestAddToPack_Duplicate verifies adding same sticker twice doesn't duplicate
func TestAddToPack_Duplicate(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	sticker := testSticker("sha256:abc123")
	if err := AddSticker(tmpDir, sticker); err != nil {
		t.Fatalf("Failed to add sticker: %v", err)
	}

	if err := CreatePack(tmpDir, "favourites", "My Favourites"); err != nil {
		t.Fatalf("Failed to create pack: %v", err)
	}

	// Add sticker twice
	if err := AddToPack(tmpDir, "favourites", []string{"sha256:abc123"}); err != nil {
		t.Fatalf("Failed to add sticker first time: %v", err)
	}
	if err := AddToPack(tmpDir, "favourites", []string{"sha256:abc123"}); err != nil {
		t.Fatalf("Failed to add sticker second time: %v", err)
	}

	pack, err := GetPack(tmpDir, "favourites")
	if err != nil {
		t.Fatalf("Failed to get pack: %v", err)
	}

	if len(pack.StickerIDs) != 1 {
		t.Errorf("Expected pack to contain 1 sticker (no duplicates), got %d", len(pack.StickerIDs))
	}
}

// TestRemoveFromPack_Success verifies removing sticker from pack
func TestRemoveFromPack_Success(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	sticker := testSticker("sha256:abc123")
	if err := AddSticker(tmpDir, sticker); err != nil {
		t.Fatalf("Failed to add sticker: %v", err)
	}

	if err := CreatePack(tmpDir, "favourites", "My Favourites"); err != nil {
		t.Fatalf("Failed to create pack: %v", err)
	}

	if err := AddToPack(tmpDir, "favourites", []string{"sha256:abc123"}); err != nil {
		t.Fatalf("Failed to add sticker to pack: %v", err)
	}

	// Remove sticker
	if err := RemoveFromPack(tmpDir, "favourites", []string{"sha256:abc123"}); err != nil {
		t.Fatalf("Failed to remove sticker from pack: %v", err)
	}

	// Verify pack is empty
	pack, err := GetPack(tmpDir, "favourites")
	if err != nil {
		t.Fatalf("Failed to get pack: %v", err)
	}
	if len(pack.StickerIDs) != 0 {
		t.Errorf("Expected pack to be empty after removal, got %d stickers", len(pack.StickerIDs))
	}

	// Verify sticker no longer references pack
	retrieved, err := GetSticker(tmpDir, "sha256:abc123")
	if err != nil {
		t.Fatalf("Failed to get sticker: %v", err)
	}
	if containsString(retrieved.InPacks, "favourites") {
		t.Error("Expected sticker to not be in 'favourites' pack after removal")
	}
}

// TestRemoveFromPack_NotInPack verifies removing sticker that isn't in pack is safe
func TestRemoveFromPack_NotInPack(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	sticker := testSticker("sha256:abc123")
	if err := AddSticker(tmpDir, sticker); err != nil {
		t.Fatalf("Failed to add sticker: %v", err)
	}

	if err := CreatePack(tmpDir, "favourites", "My Favourites"); err != nil {
		t.Fatalf("Failed to create pack: %v", err)
	}

	// Remove sticker that was never added - should not error
	if err := RemoveFromPack(tmpDir, "favourites", []string{"sha256:abc123"}); err != nil {
		t.Fatalf("Unexpected error removing sticker not in pack: %v", err)
	}
}

// TestUpdatePublished verifies tracking published rooms
func TestUpdatePublished(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if err := CreatePack(tmpDir, "favourites", "My Favourites"); err != nil {
		t.Fatalf("Failed to create pack: %v", err)
	}

	roomID := "!test:matrix.org"
	stateKey := "im.ponies.room_emotes.favourites"

	if err := UpdatePublished(tmpDir, "favourites", roomID, stateKey); err != nil {
		t.Fatalf("Failed to update published: %v", err)
	}

	pack, err := GetPack(tmpDir, "favourites")
	if err != nil {
		t.Fatalf("Failed to get pack: %v", err)
	}

	if pack.PublishedRooms[roomID] != stateKey {
		t.Errorf("Expected state key %s for room %s, got %s", stateKey, roomID, pack.PublishedRooms[roomID])
	}
}

// Helper functions

func setupTestDir(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "stickerbook-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return tmpDir
}

func testSticker(id string) Sticker {
	return Sticker{
		ID:               id,
		CollectedAt:      time.Now(),
		SourceRoom:       "!test:matrix.org",
		SourceEvent:      "$event123",
		SourceMXC:        "mxc://matrix.org/original",
		LocalMXC:         "mxc://local.org/rehosted",
		MimeType:         "image/png",
		Width:            512,
		Height:           512,
		SizeBytes:        12345,
		OriginalBody:     "test sticker",
		GeneratedAltText: "A test sticker image",
		InPacks:          []string{},
	}
}

func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
