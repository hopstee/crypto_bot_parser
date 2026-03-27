// Package app manages the core application lifecycle.
// It handles authentication, session management, HTTP client setup, and
// starting the CryptoBot websocket connection.

package app

import (
	"context"
	"crypto_bot_parser/internal/tgbot"
	"crypto_bot_parser/pkg/logger"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/adrg/xdg"
)

type App struct {
	// Application context for cancellation
	ctx context.Context
	// Cancel function for context
	cancel context.CancelFunc

	// WaitGroup to manage background tasks
	wg sync.WaitGroup

	// telegram bot instance
	bot *tgbot.Bot

	// Application name (used for directories)
	appName string
	// Data directory path
	dataDir string
}

// -bot_id=1559501630 -code=123456789 -token=8226402986:AAGKEEZD8ULDwMHa-LJj13-pL3w71lkj3Hk
// for local 8253015454:AAFDVdReBW4wQDu5gCBiiAJXaFLPZZg73yw

// New creates a new App instance.
// Parameters:
//
//	rootCtx: parent context for cancellation
//	appName: application name for data storage
func New(rootCtx context.Context, appName string) (*App, error) {
	botID := flag.String("bot_id", "", "crupto bot id")
	secreteCode := flag.String("code", "", "secrete code for authentication in bot")
	token := flag.String("token", "", "telegram bot token")
	flag.Parse()

	if *botID == "" {
		return nil, errors.New("bot id is required")
	}

	if *secreteCode == "" {
		return nil, errors.New("secret is required")
	}

	if *token == "" {
		return nil, errors.New("token is required")
	}

	ctx, cancel := context.WithCancel(rootCtx)

	a := &App{
		ctx:     ctx,
		cancel:  cancel,
		appName: appName,
	}

	if err := a.initDirs(); err != nil {
		return nil, fmt.Errorf("failed to initialize directories: %w", err)
	}

	if err := logger.NewLogger(logger.LoggerOptions{
		Mode: logger.ModeDev,
		File: filepath.Join(a.dataDir, "app.log"),
	}); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	a.bot = tgbot.NewBot(*botID, *secreteCode, a.dataDir, &a.wg)
	if err := a.bot.Init(*token); err != nil {
		return nil, fmt.Errorf("failed to initialize tg bot: %w", err)
	}

	return a, nil
}

// Run starts the app, runs authentication, and launches the CryptoBot websocket.
// It blocks until the context is canceled.
func (a *App) Run() error {
	go a.bot.Run(a.ctx)

	<-a.ctx.Done()

	return a.shutdown()
}

// initDirs creates required directories for the app (data storage, logs).
func (a *App) initDirs() error {
	a.dataDir = filepath.Join(xdg.DataHome, a.appName)

	for _, dir := range []string{a.dataDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// shutdown cancels the context, waits for background tasks, saves session, and cleans up.
func (a *App) shutdown() error {
	slog.Info("Shutdown started")

	a.cancel()

	a.wg.Wait()

	slog.Info("Shutdown completed")
	return nil
}
