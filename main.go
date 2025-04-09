package main

import (
	"context"
	"fmt"
	"os"

	"tg-dl/config"
	"tg-dl/tg_client"
	"tg-dl/types"
	"tg-dl/utils/download"
	"tg-dl/utils/list"
	"tg-dl/utils/sync"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  init                          - Create default config.json")
		fmt.Println("  list <path>                   - List available channels and select one")
		fmt.Println("  sync <path>      - Fetch messages and download media files")
		os.Exit(1)
	}

	command := os.Args[1]

	// Handle init command separately as it doesn't need client
	if command == "init" {
		if err := config.HandleInit(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Println("provide a path to download the files to")
		os.Exit(1)
	}

	client := tg_client.InitClientWithConfig(cfg)
	ctx := context.Background()
	ctx = context.WithValue(ctx, types.CtxConfigKey, cfg)
	ctx = context.WithValue(ctx, types.CtxDownloadPathKey, os.Args[2])

	if err := client.Run(ctx, func(ctx context.Context) error {
		if err := tg_client.MaybeAuth(ctx, client); err != nil {
			return fmt.Errorf("error authenticating: %w", err)
		}

		switch command {
		case "list":
			if err := list.ListChannels(ctx, client); err != nil {
				return fmt.Errorf("error listing channels: %w", err)
			}

		case "sync":
			if len(os.Args) < 3 {
				return fmt.Errorf("usage: sync <path> [--force]")
			}

			// Check for force flag
			force := false
			if len(os.Args) > 3 && os.Args[3] == "--force" {
				force = true
			}

			fmt.Println("Syncing messages...")
			if err := sync.SyncMessages(ctx, client, force); err != nil {
				return fmt.Errorf("error syncing messages: %w", err)
			}

			fmt.Println("\nDownloading media...")
			if err := download.DownloadMedia(ctx, client); err != nil {
				return fmt.Errorf("error downloading media: %w", err)
			}

		default:
			return fmt.Errorf("unknown command: %s", command)
		}

		return nil
	}); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
