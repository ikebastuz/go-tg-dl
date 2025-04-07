package download

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gotd/td/telegram"

	"tg-dl/constants"
)

func DownloadMedia(ctx context.Context, client *telegram.Client, basePath string) error {
	data, err := loadMessages(basePath)
	if err != nil {
		return fmt.Errorf("error loading messages: %w", err)
	}

	totalMediaFiles := 0
	remainingFiles := 0

	// Ensure data directory exists
	dataPath := filepath.Join(basePath, "data")
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return fmt.Errorf("error creating data directory: %w", err)
	}

	// Pre-scan files and build a map for quick lookup
	existingFiles := make(map[string]bool)
	for _, msg := range data.Messages {
		if msg.IsPhoto || msg.IsVideo {
			totalMediaFiles++
			date, _ := time.Parse(time.RFC3339, msg.Date)
			yearDir := filepath.Join(dataPath, fmt.Sprintf("%d", date.Year()))
			monthDir := filepath.Join(yearDir, fmt.Sprintf("%02d", date.Month()))

			var filename string
			if msg.IsPhoto {
				filename = filepath.Join(monthDir, fmt.Sprintf("%d.%s", msg.ID, constants.PhotoExtension))
			} else if msg.IsVideo {
				filename = filepath.Join(monthDir, fmt.Sprintf("%d.%s", msg.ID, constants.VideoExtension))
			} else {
				continue
			}

			if isFileDownloaded(filename) {
				existingFiles[filename] = true
			} else {
				remainingFiles++
			}
		}
	}

	fmt.Printf("Found %d new media files to download (out of %d total)\n", remainingFiles, totalMediaFiles)

	state, err := initializeDownloadState(client, data.ChannelID, data.AccessHash, basePath, totalMediaFiles)
	if err != nil {
		return fmt.Errorf("error initializing download state: %w", err)
	}
	defer state.mediaBar.Close()

	// Set initial progress for already completed files
	if len(existingFiles) > 0 {
		state.mediaBar.Add(len(existingFiles))
	}

	// Process each message that needs downloading
	for _, msg := range data.Messages {
		if !msg.IsPhoto && !msg.IsVideo {
			continue
		}

		date, _ := time.Parse(time.RFC3339, msg.Date)
		yearDir := filepath.Join(dataPath, fmt.Sprintf("%d", date.Year()))
		monthDir := filepath.Join(yearDir, fmt.Sprintf("%02d", date.Month()))

		var filename string
		if msg.IsPhoto {
			filename = filepath.Join(monthDir, fmt.Sprintf("%d.%s", msg.ID, constants.PhotoExtension))
		} else if msg.IsVideo {
			filename = filepath.Join(monthDir, fmt.Sprintf("%d.%s", msg.ID, constants.VideoExtension))
		} else {
			continue
		}

		if existingFiles[filename] {
			continue
		}

		if err := state.processMediaMessage(ctx, msg); err != nil {
			fmt.Printf("\nError processing message %d: %v\n", msg.ID, err)
		}

		time.Sleep(2 * time.Second) // Rate limiting
	}

	fmt.Printf("\nDownloaded media files to %s\n", state.dataPath)
	return nil
}
