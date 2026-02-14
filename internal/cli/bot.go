package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/liminalpurple/matrix-stickerbook/internal/bot"
	"github.com/liminalpurple/matrix-stickerbook/internal/config"
	"github.com/liminalpurple/matrix-stickerbook/internal/llm"
	"github.com/liminalpurple/matrix-stickerbook/internal/matrix"
	"github.com/spf13/cobra"
)

// NewBotCmd creates the bot command
func NewBotCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bot",
		Short: "Run the sticker collection bot",
		Long: `Run the Matrix bot that watches for reaction commands and collects stickers.

The bot monitors all rooms for reactions from your user account. When it detects
a !yoink, !nom, or !grab reaction, it:

  1. Downloads the image from the source homeserver
  2. Re-uploads it to your local homeserver (rehosting)
  3. Generates alt-text using Claude vision API
  4. Saves the sticker to your collection
  5. Redacts the reaction to confirm collection

The bot runs until interrupted with Ctrl+C.`,
		RunE: runBot,
	}
}

func runBot(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Verify required configuration
	if cfg.Matrix.AccessToken == "" {
		return fmt.Errorf("no access token configured - run 'stickerbook login' first")
	}
	if cfg.Anthropic.APIKey == "" {
		return fmt.Errorf("no Anthropic API key configured - set ANTHROPIC_API_KEY or add to config.yaml")
	}

	log.Println("Creating Matrix client...")
	matrixClient, err := matrix.NewClient(
		cfg.Matrix.Homeserver,
		cfg.Matrix.UserID,
		cfg.Matrix.AccessToken,
	)
	if err != nil {
		return fmt.Errorf("failed to create Matrix client: %w", err)
	}

	// Verify connection
	ctx := context.Background()
	if err := matrixClient.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to Matrix: %w", err)
	}

	log.Printf("Connected as %s", matrixClient.UserID)

	// Create LLM client
	log.Println("Creating LLM client...")
	llmClient := llm.NewClient(
		cfg.Anthropic.APIKey,
		cfg.Anthropic.Model,
		cfg.Anthropic.MaxTokens,
	)

	log.Printf("Using model: %s (max tokens: %d)", llmClient.Model(), llmClient.MaxTokens())

	// Create bot
	log.Println("Starting bot...")
	stickerbookBot := bot.NewBot(matrixClient, llmClient, cfg)

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run bot in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- stickerbookBot.Run()
	}()

	// Wait for either error or shutdown signal
	select {
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("bot error: %w", err)
		}
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
		stickerbookBot.Stop()
		// Wait for bot to finish
		if err := <-errChan; err != nil {
			return fmt.Errorf("bot shutdown error: %w", err)
		}
	}

	log.Println("Bot stopped")
	return nil
}
