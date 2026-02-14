package main

import (
	"fmt"
	"os"

	"github.com/liminalpurple/matrix-stickerbook/internal/cli"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "stickerbook",
		Short: "Matrix sticker collection and curation tool",
		Long: `Matrix Stickerbook - Bot and CLI for collecting and curating Matrix stickers.

Collect stickers by reacting with !yoink, !nom, or !grab in Matrix.
Organise collected stickers into curated packs.
Publish packs to Matrix rooms as MSC2545 state events.`,
		Version: version,
	}

	// Add commands
	rootCmd.AddCommand(cli.NewLoginCmd())
	rootCmd.AddCommand(cli.NewTestCmd())
	rootCmd.AddCommand(cli.NewBotCmd())

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
