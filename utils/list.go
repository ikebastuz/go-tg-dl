package utils

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"tg-dl/config"
	"tg-dl/constants"
	"tg-dl/types"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func ListChannels(ctx context.Context, client *telegram.Client, cfg *config.Config, basePath string) error {
	dialogs, err := client.API().MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      cfg.ChannelListSize,
	})
	if err != nil {
		return fmt.Errorf("failed to get dialogs: %w", err)
	}

	var channels []*tg.Channel
	switch d := dialogs.(type) {
	case *tg.MessagesDialogs:
		for _, chat := range d.Chats {
			if channel, ok := chat.(*tg.Channel); ok {
				channels = append(channels, channel)
			}
		}
	case *tg.MessagesDialogsSlice:
		for _, chat := range d.Chats {
			if channel, ok := chat.(*tg.Channel); ok {
				channels = append(channels, channel)
			}
		}
	default:
		return fmt.Errorf("unexpected dialogs type: %T", dialogs)
	}

	if len(channels) == 0 {
		fmt.Println("\nNo channels found")
		return nil
	}

	fmt.Println("\nAvailable channels:")
	for i := 0; i < len(channels); i++ {
		fmt.Printf("[%d]\tID: %d\t%s\n", i+1, channels[i].ID, channels[i].Title)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\nEnter the number of the channel to select (or 'q' to quit): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "q" {
			return nil
		}

		idx, err := strconv.Atoi(input)
		if err != nil || idx < 1 || idx > len(channels) {
			fmt.Println("Invalid selection. Please try again.")
			continue
		}

		selectedChannel := channels[idx-1]
		fmt.Printf("\nSelected channel: %s (ID: %d)\n", selectedChannel.Title, selectedChannel.ID)

		// Create data.json if it doesn't exist
		if err := os.MkdirAll(basePath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", basePath, err)
		}

		dataPath := filepath.Join(basePath, constants.DataPath)
		if _, err := os.Stat(dataPath); os.IsNotExist(err) {
			data := types.DataContainer{
				AccessHash: selectedChannel.AccessHash,
				Messages:   []types.MessageData{},
			}

			jsonData, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal data: %w", err)
			}

			if err := os.WriteFile(dataPath, jsonData, 0644); err != nil {
				return fmt.Errorf("failed to create data.json: %w", err)
			}
			fmt.Printf("Created %s with access hash for channel %s\n", dataPath, selectedChannel.Title)
		}
		return nil
	}
}
