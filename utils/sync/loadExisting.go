package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"tg-dl/constants"
	"tg-dl/types"
)

func loadExistingMessages(basePath string, force bool) (*syncState, error) {
	state := &syncState{
		basePath: basePath,
		force:    force,
	}

	jsonPath := filepath.Join(basePath, constants.DataPath)
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("data.json not found in %s - please run 'list %s' first to select a channel", basePath, basePath)
		}
		return nil, fmt.Errorf("failed to read data.json: %w", err)
	}

	if err := json.Unmarshal(data, &state.data); err != nil {
		return nil, fmt.Errorf("failed to parse existing data.json: %w", err)
	}

	if state.data.ChannelID == 0 {
		return nil, fmt.Errorf("invalid data.json: missing channel ID - please run 'list %s' to select a channel", basePath)
	}

	state.channelID = state.data.ChannelID

	if force {
		fmt.Println("Force flag set - fetching all messages from scratch")
		state.data.Messages = []types.MessageData{}
		return state, nil
	}

	// Find the latest message ID
	for _, msg := range state.data.Messages {
		if msg.ID > state.latestID {
			state.latestID = msg.ID
		}
	}

	return state, nil
}
