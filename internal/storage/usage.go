package storage

import (
	"fmt"
	"regexp"
	"strings"
)

// ParseUsage converts a user-friendly usage string into the canonical []string format
// Accepts: "sticker", "emoticon", "emoji" (synonym for emoticon), "both", "reset"
// Returns: []string{"sticker"}, []string{"emoticon"}, []string{"sticker", "emoticon"}, or nil (for reset)
func ParseUsage(input string) ([]string, error) {
	input = strings.ToLower(strings.TrimSpace(input))

	switch input {
	case "sticker":
		return []string{"sticker"}, nil
	case "emoticon", "emoji":
		return []string{"emoticon"}, nil
	case "both":
		return []string{"sticker", "emoticon"}, nil
	case "reset":
		return nil, nil
	default:
		return nil, fmt.Errorf("invalid usage type: %s (valid: sticker, emoticon, emoji, both, reset)", input)
	}
}

// FormatUsage converts a usage []string into a human-readable string for display
func FormatUsage(usage []string) string {
	if len(usage) == 0 {
		return "(not set)"
	}

	// Check for both
	hasSticker := false
	hasEmoticon := false
	for _, u := range usage {
		if u == "sticker" {
			hasSticker = true
		}
		if u == "emoticon" {
			hasEmoticon = true
		}
	}

	if hasSticker && hasEmoticon {
		return "both"
	}
	if hasSticker {
		return "sticker"
	}
	if hasEmoticon {
		return "emoticon"
	}

	return strings.Join(usage, ", ")
}

// ValidateShortcode checks if a shortcode name is valid for emoji usage
// Valid shortcodes: alphanumeric, underscores, hyphens, 1-64 chars
func ValidateShortcode(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("shortcode cannot be empty")
	}
	if len(name) > 64 {
		return fmt.Errorf("shortcode too long (max 64 characters)")
	}

	// Allow alphanumeric, underscore, hyphen
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(name) {
		return fmt.Errorf("shortcode must contain only letters, numbers, underscores, and hyphens")
	}

	return nil
}
