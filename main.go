package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"tg-dl/config"
	"tg-dl/tg_client"
	"tg-dl/utils"
	"tg-dl/utils/download"
	"tg-dl/utils/sync"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  init                          - Create default config.json")
		fmt.Println("  list <path>                   - List available channels and select one")
		fmt.Println("  sync <channel_id> <path>      - Fetch messages and download media files")
		fmt.Println("  purge <path>                  - Clean download tracking data")
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

	client := tg_client.InitClientWithConfig(cfg)
	ctx := context.Background()

	if err := client.Run(ctx, func(ctx context.Context) error {
		if err := tg_client.MaybeAuth(ctx, client, cfg); err != nil {
			return fmt.Errorf("error authenticating: %w", err)
		}

		switch command {
		case "list":
			if len(os.Args) < 3 {
				return fmt.Errorf("usage: list <path>")
			}
			if err := utils.ListChannels(ctx, client, cfg, os.Args[2]); err != nil {
				return fmt.Errorf("error listing channels: %w", err)
			}

		case "sync":
			if len(os.Args) < 4 {
				return fmt.Errorf("usage: sync <channel_id> <path> [--force]")
			}
			channelID, err := strconv.ParseInt(os.Args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid channel ID: %w", err)
			}

			// Check for force flag
			force := false
			if len(os.Args) > 4 && os.Args[4] == "--force" {
				force = true
			}

			// Use configured download path if none specified
			downloadPath := os.Args[3]

			fmt.Println("Syncing messages...")
			if err := sync.SyncMessages(ctx, client, channelID, downloadPath, force, cfg); err != nil {
				return fmt.Errorf("error syncing messages: %w", err)
			}

			fmt.Println("\nDownloading media...")
			if err := download.DownloadMedia(ctx, client, channelID, downloadPath); err != nil {
				return fmt.Errorf("error downloading media: %w", err)
			}

		case "purge":
			if len(os.Args) != 3 {
				return fmt.Errorf("usage: purge <path>")
			}
			if err := utils.PurgeData(ctx, client, os.Args[2]); err != nil {
				return fmt.Errorf("error purging data: %w", err)
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
