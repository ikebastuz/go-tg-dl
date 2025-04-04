package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"tg-dl/constants"
	"tg-dl/types"
)

func saveMessages(data types.DataContainer, basePath string) error {
	jsonPath := filepath.Join(basePath, constants.DataPath)
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %w", err)
	}

	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}
	return nil
}
