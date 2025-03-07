// Package main is the entry point for the wampa command
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"wampa/pkg/wampa"
)

func main() {
	// Create a context that can be cancelled by signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals for graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalCh
		log.Println("Received shutdown signal")
		cancel()
	}()

	if err := wampa.Run(ctx, os.Args[1:]); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
