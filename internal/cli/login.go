// Package cli provides command-line interface commands for stickerbook.
package cli

import (
	"fmt"

	"github.com/liminalpurple/matrix-stickerbook/internal/auth"
	"github.com/liminalpurple/matrix-stickerbook/internal/config"
	"github.com/spf13/cobra"
)

// NewLoginCmd creates the login command
func NewLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Matrix homeserver",
		Long: `Interactive login to Matrix homeserver.

Prompts for homeserver URL, user ID, and password, then saves credentials
to the configuration file for future use.`,
		RunE: runLogin,
	}
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Perform interactive login
	fmt.Println("Matrix Stickerbook - Login")
	fmt.Println()

	creds, err := auth.InteractiveLogin()
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	fmt.Println()
	fmt.Println("Login successful!")
	fmt.Printf("User ID: %s\n", creds.UserID)
	fmt.Printf("Device ID: %s\n", creds.DeviceID)
	fmt.Println()

	// Load existing config (or create new)
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Update Matrix settings
	cfg.Matrix.Homeserver = creds.Homeserver
	cfg.Matrix.UserID = creds.UserID
	cfg.Matrix.DeviceID = creds.DeviceID
	cfg.Matrix.AccessToken = creds.AccessToken

	// Save config
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	configDir, _ := config.GetConfigDir()
	fmt.Printf("Credentials saved to: %s/config.yaml\n", configDir)
	fmt.Println()
	fmt.Println("You can now run 'stickerbook bot' to start collecting stickers!")

	return nil
}
