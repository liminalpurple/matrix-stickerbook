// Package auth provides Matrix authentication functionality.
package auth

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

// LoginCredentials holds the result of a successful login
type LoginCredentials struct {
	Homeserver  string
	UserID      string
	DeviceID    string
	AccessToken string
}

// InteractiveLogin prompts the user for credentials and performs Matrix login
func InteractiveLogin() (*LoginCredentials, error) {
	reader := bufio.NewReader(os.Stdin)

	// Prompt for homeserver
	fmt.Print("Homeserver URL (e.g., https://matrix.org): ")
	homeserver, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read homeserver: %w", err)
	}
	homeserver = strings.TrimSpace(homeserver)

	// Prompt for user ID
	fmt.Print("User ID (e.g., @morgan:matrix.org): ")
	userID, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read user ID: %w", err)
	}
	userID = strings.TrimSpace(userID)

	// Prompt for password (hidden input)
	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Print newline after password input
	if err != nil {
		return nil, fmt.Errorf("failed to read password: %w", err)
	}
	password := string(passwordBytes)

	// Create Matrix client
	client, err := mautrix.NewClient(homeserver, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to create Matrix client: %w", err)
	}

	// Perform login
	resp, err := client.Login(context.Background(), &mautrix.ReqLogin{
		Type: mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{
			Type: mautrix.IdentifierTypeUser,
			User: userID,
		},
		Password:                 password,
		DeviceID:                 id.DeviceID("STICKERBOOK"),
		InitialDeviceDisplayName: "Matrix Stickerbook",
	})
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return &LoginCredentials{
		Homeserver:  homeserver,
		UserID:      resp.UserID.String(),
		DeviceID:    resp.DeviceID.String(),
		AccessToken: resp.AccessToken,
	}, nil
}
