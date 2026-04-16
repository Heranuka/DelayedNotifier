// Package main provides the entry point for the Backforge server.
//
// It initializes the application, sets up signal handling for graceful shutdown,
// and runs the HTTP server.
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"delay/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := app.New(ctx)
	if err != nil {
		log.Printf("failed to create app: %v", err)
		return
	}

	a.Logger.Info("Starting notification worker...")
	go func() {
		if err := a.Worker.Run(ctx); err != nil && err != context.Canceled {
			a.Logger.Error("worker stopped with error", "err", err)
		} else {
			a.Logger.Info("worker stopped gracefully")
		}
	}()

	if err := a.Run(ctx); err != nil {
		a.Logger.Error("app stopped with error", "err", err)
	}
}
