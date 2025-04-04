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
		return state, nil // Not an error, just no existing data
	}

	if err := json.Unmarshal(data, &state.data); err != nil {
		return nil, fmt.Errorf("failed to parse existing data.json: %w", err)
	}

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
