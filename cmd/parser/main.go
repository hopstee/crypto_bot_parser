// Package main is the entry point of the Crypto Bot Parser application.
// It sets up signal handling and starts the app.

package main

import (
	"context"
	"crypto_bot_parser/internal/app"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	rand.New(rand.NewSource(time.Now().UnixNano()))

	start(ctx)
}

// start initializes and runs the application.
// Parameters:
//
//	ctx: context for graceful shutdown
func start(ctx context.Context) {
	a, err := app.New(ctx, "cryptoBotParser")
	if err != nil {
		slog.Error("Failed to create app", slog.Any("error", err.Error()))
		os.Exit(1)
	}

	if err := a.Run(); err != nil {
		slog.Error("Failed to run app", slog.Any("error", err.Error()))
		os.Exit(1)
	}
}
