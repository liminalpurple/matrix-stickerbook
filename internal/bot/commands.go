package bot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/liminalpurple/matrix-stickerbook/internal/storage"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// handleMessage processes text messages looking for !sticker commands
func (b *Bot) handleMessage(ctx context.Context, evt *event.Event) {
	// Only process messages from our user
	if evt.Sender != b.client.UserID {
		return
	}

	// Parse message content
	content, ok := evt.Content.Parsed.(*event.MessageEventContent)
	if !ok {
		return
	}

	// Skip edits (don't respond to our own command results)
	if content.RelatesTo != nil && content.RelatesTo.Type == event.RelReplace {
		return
	}

	// Only process text messages
	if content.MsgType != event.MsgText {
		return
	}

	// Check if message starts with !sticker (with or without space)
	body := strings.TrimSpace(content.Body)
	if !strings.HasPrefix(body, "!sticker") {
		return
	}

	log.Printf("Processing command: %s", body)

	// Parse and execute command
	result := b.executeCommand(ctx, body)

	// Edit the original message with the result
	if err := b.editMessage(ctx, evt.RoomID, evt.ID, body, result); err != nil {
		log.Printf("Error editing message: %v", err)
	}
}

// showHelp returns a help message with all available commands
func (b *Bot) showHelp() string {
	return "Pack Management:\n\n" +
		"- !sticker pack list - List all packs with sticker counts\n" +
		"- !sticker pack create <name> - Create a new pack\n" +
		"- !sticker pack show <pack> - Show stickers in a pack\n" +
		"- !sticker pack add <pack> <sticker-id> - Add sticker to pack\n" +
		"- !sticker pack remove <pack> <sticker-id> - Remove sticker from pack\n" +
		"- !sticker pack avatar <pack> <mxc-uri> - Set pack icon\n" +
		"- !sticker pack usage <pack> <type> - Set default usage (sticker/emoticon/both/reset)\n" +
		"- !sticker pack publish <pack> [room-id] - Publish to room (or all saved)\n\n" +
		"Listing:\n\n" +
		"- !sticker list unsorted - Show stickers not in any pack\n" +
		"- !sticker show <sticker-id> - Show sticker with metadata and image\n\n" +
		"Management:\n\n" +
		"- !sticker name <sticker-id> <shortcode> - Set emoji shortcode (e.g., happy_cat)\n" +
		"- !sticker usage <sticker-id> <type> - Set usage (sticker/emoticon/both/reset)\n" +
		"- !sticker delete <sticker-id> - Delete sticker from collection\n\n" +
		"**React to any sticker with `!yoink`, `!nom`, or `!grab` to collect it!**"
}

// executeCommand parses and executes a !sticker command
func (b *Bot) executeCommand(ctx context.Context, body string) string {
	// Remove "!sticker" prefix (handle both "!sticker" and "!sticker ...")
	body = strings.TrimSpace(body)

	// Show help if just "!sticker" with no args
	if len(body) <= 8 { // "!sticker" is 8 chars
		return b.showHelp()
	}

	// Parse args (skip "!sticker ")
	args := strings.Fields(body[8:])
	if len(args) == 0 {
		return b.showHelp()
	}

	switch args[0] {
	case "pack":
		return b.handlePackCommand(args[1:])
	case "list":
		return b.handleListCommand(args[1:])
	case "show":
		if len(args) < 2 {
			return "❌ Usage: !sticker show <sticker-id>"
		}
		return b.stickerShow(args[1])
	case "delete", "remove":
		if len(args) < 2 {
			return "❌ Usage: !sticker delete <sticker-id>"
		}
		return b.stickerDelete(args[1])
	case "usage":
		if len(args) < 3 {
			return "❌ Usage: !sticker usage <sticker-id> <sticker|emoticon|emoji|both|reset>\n\nSets how this sticker can be used. Use 'reset' to clear override and inherit from pack."
		}
		return b.stickerUsage(args[1], args[2])
	case "name":
		if len(args) < 3 {
			return "❌ Usage: !sticker name <sticker-id> <shortcode>\n\nSets the emoji shortcode name (e.g., 'happy_cat' becomes :happy_cat:). Defaults to SHA256 hash."
		}
		return b.stickerName(args[1], args[2])
	default:
		return fmt.Sprintf("❌ Unknown command: %s\n\n%s", args[0], b.showHelp())
	}
}

// handlePackCommand handles !sticker pack <subcommand>
func (b *Bot) handlePackCommand(args []string) string {
	if len(args) == 0 {
		return "❌ No pack subcommand specified. Try: pack list, pack create, pack add, pack remove, pack show, pack avatar, pack publish"
	}

	switch args[0] {
	case "list":
		return b.packList()
	case "create":
		if len(args) < 2 {
			return "❌ Usage: !sticker pack create <name>"
		}
		// Join all remaining args as pack name
		packName := strings.Join(args[1:], " ")
		return b.packCreate(packName)
	case "add":
		if len(args) < 3 {
			return "❌ Usage: !sticker pack add <pack-name> <sticker-id>\n\nExample: !sticker pack add favourites abc123...\n\nUse `!sticker pack list` to see available packs, or create one with `!sticker pack create <name>`"
		}
		return b.packAdd(args[1], args[2])
	case "remove":
		if len(args) < 3 {
			return "❌ Usage: !sticker pack remove <pack-name> <sticker-id>\n\nExample: !sticker pack remove favourites abc123..."
		}
		return b.packRemove(args[1], args[2])
	case "show":
		if len(args) < 2 {
			return "❌ Usage: !sticker pack show <pack>"
		}
		return b.packShow(args[1])
	case "publish":
		if len(args) < 2 {
			return "❌ Usage: !sticker pack publish <pack-name> [room-id]\n\nPublish to a specific room: !sticker pack publish favourites !roomid:matrix.org\nRe-publish to all saved rooms: !sticker pack publish favourites"
		}
		// Optional room ID - if not provided, republish to all saved rooms
		roomID := ""
		if len(args) >= 3 {
			roomID = args[2]
		}
		return b.packPublish(args[1], roomID)
	case "avatar":
		if len(args) < 3 {
			return "❌ Usage: !sticker pack avatar <pack-name> <mxc-uri>\n\nExample: !sticker pack avatar favourites mxc://matrix.org/abc123..."
		}
		return b.packAvatar(args[1], args[2])
	case "usage":
		if len(args) < 3 {
			return "❌ Usage: !sticker pack usage <pack-name> <sticker|emoticon|emoji|both|reset>\n\nSets default usage for all stickers in this pack. Individual stickers can override this."
		}
		return b.packUsage(args[1], args[2])
	default:
		return fmt.Sprintf("❌ Unknown pack subcommand: %s", args[0])
	}
}

// handleListCommand handles !sticker list <subcommand>
func (b *Bot) handleListCommand(args []string) string {
	if len(args) == 0 {
		return "❌ No list subcommand specified. Try: list unsorted"
	}

	switch args[0] {
	case "unsorted":
		return b.listUnsorted()
	default:
		return fmt.Sprintf("❌ Unknown list subcommand: %s", args[0])
	}
}

// packList lists all packs with sticker counts
func (b *Bot) packList() string {
	packs, err := storage.ListPacks(b.storageDir)
	if err != nil {
		return fmt.Sprintf("❌ Error loading packs: %v", err)
	}

	// Count unsorted stickers
	collection, err := storage.LoadCollection(b.storageDir)
	if err != nil {
		return fmt.Sprintf("❌ Error loading collection: %v", err)
	}

	unsortedCount := 0
	for _, sticker := range collection.Stickers {
		if len(sticker.InPacks) == 0 {
			unsortedCount++
		}
	}

	var result strings.Builder

	// Always show "unsorted" meta-pack (even if 0)
	result.WriteString(fmt.Sprintf("- unsorted (%d)\n", unsortedCount))

	for _, pack := range packs {
		result.WriteString(fmt.Sprintf("- %s (%d)\n", pack.Name, len(pack.StickerIDs)))
	}

	// Add helpful message if no packs created yet
	if len(packs) == 0 {
		result.WriteString("\nCreate a pack with: !sticker pack create <name>")
	}

	return result.String()
}

// packCreate creates a new pack
func (b *Bot) packCreate(name string) string {
	// Keep original name for display
	displayName := name

	// Sanitize pack name for ID (lowercase, no spaces)
	packID := strings.ToLower(strings.ReplaceAll(name, " ", "-"))

	// Forbid "unsorted" as it's a virtual pack
	if packID == "unsorted" {
		return "❌ Cannot create pack named 'unsorted' - this is a reserved name for stickers not in any pack"
	}

	// Create pack with display name and attribution
	if err := storage.CreatePackWithAttribution(b.storageDir, packID, displayName, string(b.client.UserID)); err != nil {
		return fmt.Sprintf("❌ Error creating pack: %v", err)
	}

	return fmt.Sprintf("✅ Created pack: %s", displayName)
}

// packAdd adds a sticker to a pack
func (b *Bot) packAdd(packName, stickerID string) string {
	if err := storage.AddToPack(b.storageDir, packName, []string{stickerID}); err != nil {
		return fmt.Sprintf("❌ Error adding to pack: %v", err)
	}

	return fmt.Sprintf("✅ Added sticker to pack: %s", packName)
}

// packRemove removes a sticker from a pack
func (b *Bot) packRemove(packName, stickerID string) string {
	if err := storage.RemoveFromPack(b.storageDir, packName, []string{stickerID}); err != nil {
		return fmt.Sprintf("❌ Error removing from pack: %v", err)
	}

	return fmt.Sprintf("✅ Removed sticker from pack: %s", packName)
}

// packShow shows stickers in a pack
func (b *Bot) packShow(packName string) string {
	pack, err := storage.GetPack(b.storageDir, packName)
	if err != nil {
		return fmt.Sprintf("❌ Error loading pack: %v", err)
	}

	if len(pack.StickerIDs) == 0 {
		return "Pack is empty"
	}

	var result strings.Builder

	// Load stickers to show their alt-text
	collection, err := storage.LoadCollection(b.storageDir)
	if err != nil {
		return fmt.Sprintf("❌ Error loading collection: %v", err)
	}

	for i, stickerID := range pack.StickerIDs {
		// Find the sticker in collection
		var altText, name string
		for _, sticker := range collection.Stickers {
			if sticker.ID == stickerID {
				altText = sticker.GeneratedAltText
				name = sticker.Name
				break
			}
		}

		if altText == "" {
			altText = "(no alt-text)"
		}

		// Use code formatting for ID, proper markdown ordered list
		result.WriteString(fmt.Sprintf("%d. `%s` (:%s:) - %s\n", i+1, stickerID, name, altText))
	}

	return result.String()
}

// packPublish publishes a pack to a Matrix room (or all previously published rooms if roomID is empty)
func (b *Bot) packPublish(packName, roomID string) string {
	// If no room ID provided, republish to all saved rooms
	if roomID == "" {
		pack, err := storage.GetPack(b.storageDir, packName)
		if err != nil {
			return fmt.Sprintf("❌ Error loading pack: %v", err)
		}

		if len(pack.PublishedRooms) == 0 {
			return "❌ Pack has not been published to any rooms yet\n\nUse: !sticker pack publish <pack> <room-id> to publish to a specific room"
		}

		// Publish to all saved rooms
		successCount := 0
		var errors []string
		for savedRoomID := range pack.PublishedRooms {
			if err := b.client.PublishPack(b.ctx, b.storageDir, packName, id.RoomID(savedRoomID)); err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", savedRoomID, err))
			} else {
				successCount++
			}
		}

		if len(errors) > 0 {
			return fmt.Sprintf("⚠️ Published to %d/%d rooms\n\nErrors:\n%s", successCount, len(pack.PublishedRooms), strings.Join(errors, "\n"))
		}

		return fmt.Sprintf("✅ Published pack '%s' to %d room(s)", packName, successCount)
	}

	// Validate room ID format
	if !strings.HasPrefix(roomID, "!") {
		return "❌ Invalid room ID - must start with !\n\nExample: !roomid:matrix.org"
	}

	// Publish to specific room
	if err := b.client.PublishPack(b.ctx, b.storageDir, packName, id.RoomID(roomID)); err != nil {
		return fmt.Sprintf("❌ Error publishing pack: %v", err)
	}

	return fmt.Sprintf("✅ Published pack '%s' to room %s", packName, roomID)
}

// packAvatar sets the avatar for a pack
func (b *Bot) packAvatar(packName, avatarURL string) string {
	// Validate MXC URI format
	if !strings.HasPrefix(avatarURL, "mxc://") {
		return "❌ Invalid MXC URI - must start with mxc://\n\nExample: mxc://matrix.org/abc123..."
	}

	// Set the avatar
	if err := storage.SetPackAvatar(b.storageDir, packName, avatarURL); err != nil {
		return fmt.Sprintf("❌ Error setting pack avatar: %v", err)
	}

	return fmt.Sprintf("✅ Set avatar for pack: %s", packName)
}

// stickerShow displays a sticker with metadata and image
func (b *Bot) stickerShow(stickerID string) string {
	collection, err := storage.LoadCollection(b.storageDir)
	if err != nil {
		return fmt.Sprintf("❌ Error loading collection: %v", err)
	}

	// Find the sticker
	var sticker *storage.Sticker
	for i := range collection.Stickers {
		if collection.Stickers[i].ID == stickerID {
			sticker = &collection.Stickers[i]
			break
		}
	}

	if sticker == nil {
		return fmt.Sprintf("❌ Sticker not found: %s", stickerID)
	}

	// Build metadata as markdown list
	var result strings.Builder

	// Alt-text
	altText := sticker.GeneratedAltText
	if altText == "" {
		altText = sticker.OriginalBody
	}
	if altText == "" {
		altText = "Sticker"
	}

	// Metadata as list
	result.WriteString(fmt.Sprintf("- **ID:** `%s`\n", sticker.ID))
	result.WriteString(fmt.Sprintf("- **Name:** `:%s:`\n", sticker.Name))
	result.WriteString(fmt.Sprintf("- **Alt-text:** %s\n", altText))
	result.WriteString(fmt.Sprintf("- **Size:** %dx%d, %s\n", sticker.Width, sticker.Height, sticker.MimeType))

	// Packs
	if len(sticker.InPacks) > 0 {
		result.WriteString(fmt.Sprintf("- **Packs:** %s\n", strings.Join(sticker.InPacks, ", ")))
	} else {
		result.WriteString("- **Packs:** (unsorted)\n")
	}

	// Blank line before image
	result.WriteString("\n")

	// Markdown image for clients that support it
	result.WriteString(fmt.Sprintf("![%s](%s)", altText, sticker.LocalMXC))

	return result.String()
}

// stickerDelete deletes a sticker from the collection
func (b *Bot) stickerDelete(stickerID string) string {
	if err := storage.DeleteSticker(b.storageDir, stickerID); err != nil {
		return fmt.Sprintf("❌ Error deleting sticker: %v", err)
	}

	return fmt.Sprintf("✅ Deleted sticker: %s", stickerID)
}

// listUnsorted lists stickers not in any pack
func (b *Bot) listUnsorted() string {
	collection, err := storage.LoadCollection(b.storageDir)
	if err != nil {
		return fmt.Sprintf("❌ Error loading collection: %v", err)
	}

	var unsorted []storage.Sticker
	for _, sticker := range collection.Stickers {
		if len(sticker.InPacks) == 0 {
			unsorted = append(unsorted, sticker)
		}
	}

	if len(unsorted) == 0 {
		return "All stickers are organized into packs!"
	}

	var result strings.Builder

	for i, sticker := range unsorted {
		altText := sticker.GeneratedAltText
		if altText == "" {
			altText = "(no alt-text)"
		}

		// Use code formatting for ID, proper markdown ordered list
		result.WriteString(fmt.Sprintf("%d. `%s` (:%s:) - %s\n", i+1, sticker.ID, sticker.Name, altText))
	}

	return result.String()
}

// editMessage edits a message to show the command result
func (b *Bot) editMessage(ctx context.Context, roomID id.RoomID, eventID id.EventID, originalBody, result string) error {
	// Construct the edited message body
	newBody := fmt.Sprintf("%s\n\n%s", originalBody, result)

	// Convert markdown to HTML for formatted_body
	formattedBody := markdownToHTML(newBody)

	// Create edit content
	content := &event.MessageEventContent{
		MsgType:       event.MsgText,
		Body:          newBody,
		Format:        event.FormatHTML,
		FormattedBody: formattedBody,
		NewContent: &event.MessageEventContent{
			MsgType:       event.MsgText,
			Body:          newBody,
			Format:        event.FormatHTML,
			FormattedBody: formattedBody,
		},
		RelatesTo: &event.RelatesTo{
			Type:    event.RelReplace,
			EventID: eventID,
		},
	}

	_, err := b.client.SendMessageEvent(ctx, roomID, event.EventMessage, content)
	return err
}

// stickerUsage sets the usage types for a specific sticker
func (b *Bot) stickerUsage(stickerID, usageStr string) string {
	usage, err := storage.ParseUsage(usageStr)
	if err != nil {
		return fmt.Sprintf("❌ %v", err)
	}

	if err := storage.SetStickerUsage(b.storageDir, stickerID, usage); err != nil {
		return fmt.Sprintf("❌ Error setting sticker usage: %v", err)
	}

	if usage == nil {
		return fmt.Sprintf("✅ Reset usage for sticker %s (will inherit from pack)", stickerID)
	}

	return fmt.Sprintf("✅ Set sticker %s usage to: %s", stickerID, storage.FormatUsage(usage))
}

// stickerName sets the shortcode name for a specific sticker
func (b *Bot) stickerName(stickerID, name string) string {
	// Validate shortcode format
	if err := storage.ValidateShortcode(name); err != nil {
		return fmt.Sprintf("❌ Invalid shortcode: %v", err)
	}

	if err := storage.SetStickerName(b.storageDir, stickerID, name); err != nil {
		return fmt.Sprintf("❌ Error setting sticker name: %v", err)
	}

	return fmt.Sprintf("✅ Set sticker shortcode to: :%s:", name)
}

// packUsage sets the default usage for all stickers in a pack
func (b *Bot) packUsage(packName, usageStr string) string {
	usage, err := storage.ParseUsage(usageStr)
	if err != nil {
		return fmt.Sprintf("❌ %v", err)
	}

	if err := storage.SetPackUsage(b.storageDir, packName, usage); err != nil {
		return fmt.Sprintf("❌ Error setting pack usage: %v", err)
	}

	if usage == nil {
		return fmt.Sprintf("✅ Reset usage for pack %s (will use default: both)", packName)
	}

	return fmt.Sprintf("✅ Set pack %s default usage to: %s", packName, storage.FormatUsage(usage))
}

// markdownToHTML converts markdown to HTML for Matrix formatted_body
func markdownToHTML(text string) string {
	// Create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)

	// Parse markdown
	doc := p.Parse([]byte(text))

	// Create HTML renderer
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	// Render to HTML and return as string
	return string(markdown.Render(doc, renderer))
}
