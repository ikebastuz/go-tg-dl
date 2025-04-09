package list

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

func ListChannels(ctx context.Context, client *telegram.Client) error {
	cfg, ok := ctx.Value(types.CtxConfigKey).(*config.Config)
	if !ok {
		return fmt.Errorf("config not found in context")
	}

	basePath := ctx.Value(types.CtxDownloadPathKey).(string)

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

	selectedChannel, shouldQuit := selectChannel(reader, channels)
	if shouldQuit {
		return nil
	}

	fmt.Printf("\nSelected channel: %s (ID: %d)\n", selectedChannel.Title, selectedChannel.ID)

	// Create data.json if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", basePath, err)
	}

	dataPath := filepath.Join(basePath, constants.DataPath)
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		data := types.DataContainer{
			ChannelID:  selectedChannel.ID,
			AccessHash: selectedChannel.AccessHash,
			Messages:   []types.MessageData{},
		}

		if err := saveData(&data, dataPath); err != nil {
			return fmt.Errorf("failed to save data: %w", err)
		}
	} else {
		existingData, err := loadData(dataPath)
		if err != nil {
			return fmt.Errorf("failed to load data: %w", err)
		}

		if existingData.ChannelID != selectedChannel.ID {
			overwrite, err := confirmOverwrite(reader, dataPath)
			if err != nil {
				return fmt.Errorf("error reading input: %w", err)
			}
			if overwrite {
				existingData.ChannelID = selectedChannel.ID
				existingData.AccessHash = selectedChannel.AccessHash
				existingData.Messages = []types.MessageData{}

				if err := saveData(existingData, dataPath); err != nil {
					return fmt.Errorf("failed to save data: %w", err)
				}
			} else {
				fmt.Printf("%s not overwritten\n", dataPath)
				return nil
			}
		}
		return nil
	}
	return nil
}

func selectChannel(reader *bufio.Reader, channels []*tg.Channel) (*tg.Channel, bool) {
	fmt.Print("\nEnter the number of the channel to select (or 'q' to quit): ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, true
	}

	input = strings.TrimSpace(input)
	if input == "q" {
		return nil, true
	}

	idx, err := strconv.ParseInt(input, 10, 64)
	if err != nil || idx < 1 || int(idx) > len(channels) {
		fmt.Println("Invalid selection. Please try again.")
		return nil, true
	}

	return channels[int(idx)-1], false
}

func saveData(data *types.DataContainer, path string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to create %s: %w", path, err)
	}

	fmt.Printf("Created %s with access hash for channel\n", path)
	return nil
}

func loadData(path string) (*types.DataContainer, error) {
	jsonData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	data := types.DataContainer{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return &data, nil
}

func confirmOverwrite(reader *bufio.Reader, path string) (bool, error) {
	fmt.Printf("Data.json already exists for channel %s\n", path)
	fmt.Print("Do you want to overwrite it? (y/n): ")
	overwrite, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("error reading input: %w", err)
	}

	overwrite = strings.TrimSpace(overwrite)
	if overwrite == "y" {
		return true, nil
	}
	return false, nil
}
