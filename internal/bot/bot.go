// Package bot implements the Matrix bot that watches for reaction commands.
package bot

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/liminalpurple/matrix-stickerbook/internal/config"
	"github.com/liminalpurple/matrix-stickerbook/internal/llm"
	"github.com/liminalpurple/matrix-stickerbook/internal/matrix"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// simpleStore implements a minimal mautrix.SyncStore that only tracks next_batch
type simpleStore struct {
	mu        sync.RWMutex
	nextBatch string
}

func (s *simpleStore) SaveFilterID(ctx context.Context, userID id.UserID, filterID string) error {
	return nil
}
func (s *simpleStore) LoadFilterID(ctx context.Context, userID id.UserID) (string, error) {
	return "", nil
}
func (s *simpleStore) SaveNextBatch(ctx context.Context, userID id.UserID, nextBatchToken string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextBatch = nextBatchToken
	return nil
}
func (s *simpleStore) LoadNextBatch(ctx context.Context, userID id.UserID) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nextBatch, nil
}

// Bot watches Matrix rooms for reaction commands and collects stickers
type Bot struct {
	client     *matrix.Client
	llmClient  *llm.Client
	storageDir string
	syncer     *mautrix.DefaultSyncer
	ctx        context.Context
	cancel     context.CancelFunc
	config     *config.Config
	nextBatch  string
}

// NewBot creates a new bot instance
func NewBot(matrixClient *matrix.Client, llmClient *llm.Client, cfg *config.Config) *Bot {
	ctx, cancel := context.WithCancel(context.Background())

	// Create store with initial next_batch
	store := &simpleStore{
		nextBatch: cfg.Matrix.NextBatch,
	}

	// Set store on client so it uses our next_batch
	matrixClient.Client.Store = store

	bot := &Bot{
		client:     matrixClient,
		llmClient:  llmClient,
		storageDir: cfg.Storage.DataDir,
		syncer:     matrixClient.Syncer.(*mautrix.DefaultSyncer),
		ctx:        ctx,
		cancel:     cancel,
		config:     cfg,
		nextBatch:  cfg.Matrix.NextBatch,
	}

	// Register event handlers
	bot.syncer.OnEventType(event.EventReaction, bot.handleReaction)
	bot.syncer.OnEventType(event.EventMessage, bot.handleMessage)

	return bot
}

// Run starts the bot's sync loop
func (b *Bot) Run() error {
	log.Println("Starting bot sync loop...")

	// Log resume point if we have one
	if b.nextBatch != "" {
		truncated := b.nextBatch
		if len(truncated) > 20 {
			truncated = truncated[:20] + "..."
		}
		log.Printf("Resuming from next_batch: %s", truncated)
	} else {
		log.Println("No previous sync token, starting from current state")
	}

	// Start hourly ticker to save next_batch
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// Also check every 10 seconds for first sync completion
	firstSyncCheck := time.NewTicker(10 * time.Second)
	defer firstSyncCheck.Stop()

	// Start sync loop in goroutine
	syncErr := make(chan error, 1)
	go func() {
		log.Println("Sync goroutine started, waiting for events...")
		if err := b.client.SyncWithContext(b.ctx); err != nil {
			if err != context.Canceled {
				syncErr <- err
			}
		}
		log.Println("Sync goroutine exited")
	}()

	// Handle periodic saves and shutdown
	savedFirst := false
	for {
		select {
		case <-firstSyncCheck.C:
			// Check if we have a next_batch from first sync
			if !savedFirst {
				if nb, err := b.client.Client.Store.LoadNextBatch(context.Background(), b.client.UserID); err == nil && nb != "" && nb != b.nextBatch {
					b.nextBatch = nb
					log.Printf("First sync completed, next_batch: %s", nb[:min(len(nb), 20)])
					if err := b.saveNextBatch(); err != nil {
						log.Printf("Warning: failed to save next_batch after first sync: %v", err)
					} else {
						log.Println("✅ Saved initial next_batch to config")
						savedFirst = true
						firstSyncCheck.Stop()
					}
				}
			}

		case <-ticker.C:
			// Save next_batch every hour
			if err := b.saveNextBatch(); err != nil {
				log.Printf("Warning: failed to save next_batch: %v", err)
			} else {
				log.Println("Saved next_batch checkpoint")
			}

		case err := <-syncErr:
			return fmt.Errorf("sync error: %w", err)

		case <-b.ctx.Done():
			log.Println("Bot sync loop stopped")
			return nil
		}
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Stop gracefully shuts down the bot
func (b *Bot) Stop() {
	log.Println("Stopping bot...")
	b.cancel()
	b.client.StopSync()

	// Save final next_batch on graceful shutdown (already updated by OnSync callback)
	if err := b.saveNextBatch(); err != nil {
		log.Printf("Warning: failed to save next_batch on shutdown: %v", err)
	} else {
		log.Println("Saved final next_batch on shutdown")
	}
}

// saveNextBatch persists the current next_batch token to config
func (b *Bot) saveNextBatch() error {
	// Read latest next_batch from store (updated by sync)
	if nb, err := b.client.Client.Store.LoadNextBatch(context.Background(), b.client.UserID); err == nil {
		b.nextBatch = nb
	}
	b.config.Matrix.NextBatch = b.nextBatch
	return config.Save(b.config)
}

// handleReaction is called for every m.reaction event
func (b *Bot) handleReaction(ctx context.Context, evt *event.Event) {
	// Only process reactions from our user
	if evt.Sender != b.client.UserID {
		return
	}

	log.Printf("📩 Received reaction event from %s", evt.Sender)

	// Delegate to reaction handler
	if err := b.processReaction(ctx, evt); err != nil {
		log.Printf("Error processing reaction: %v", err)
	}
}
