// Package matrix provides Matrix client functionality for stickerbook.
// It wraps mautrix-go with application-specific operations.
package matrix

import (
	"context"
	"fmt"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

// Client wraps the mautrix client with stickerbook-specific functionality
type Client struct {
	*mautrix.Client
	UserID id.UserID
}

// NewClient creates a new Matrix client
func NewClient(homeserver string, userID string, accessToken string) (*Client, error) {
	client, err := mautrix.NewClient(homeserver, id.UserID(userID), accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Matrix client: %w", err)
	}

	return &Client{
		Client: client,
		UserID: id.UserID(userID),
	}, nil
}

// Connect verifies the connection and starts syncing
func (c *Client) Connect(ctx context.Context) error {
	// Verify credentials with a whoami request
	resp, err := c.Whoami(ctx)
	if err != nil {
		return fmt.Errorf("failed to verify credentials: %w", err)
	}

	if resp.UserID != c.UserID {
		return fmt.Errorf("user ID mismatch: expected %s, got %s", c.UserID, resp.UserID)
	}

	return nil
}

// StartSync begins syncing events from the homeserver
func (c *Client) StartSync(ctx context.Context) error {
	// Start syncing - will be configured with event handlers later
	return c.Sync()
}
