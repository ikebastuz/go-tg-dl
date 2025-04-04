package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"tg-dl/constants"
)

func loadExistingMessages(basePath string, force bool) (*syncState, error) {
	state := &syncState{
		basePath: basePath,
		force:    force,
	}

	if force {
		fmt.Println("Force flag set - fetching all messages from scratch")
		return state, nil
	}

	jsonPath := filepath.Join(basePath, constants.DataPath)
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return state, nil // Not an error, just no existing data
	}

	if err := json.Unmarshal(data, &state.data); err != nil {
		return nil, fmt.Errorf("failed to parse existing data.json: %w", err)
	}

	// Find the latest message ID
	for _, msg := range state.data.Messages {
		if msg.ID > state.latestID {
			state.latestID = msg.ID
		}
	}

	return state, nil
}
