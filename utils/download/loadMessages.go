package download

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"tg-dl/constants"
	"tg-dl/types"
)

func loadMessages(basePath string) (*types.DataContainer, error) {
	jsonPath := filepath.Join(basePath, constants.DataPath)
	data_raw, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read data.json: %w", err)
	}

	var data types.DataContainer
	if err := json.Unmarshal(data_raw, &data); err != nil {
		return nil, fmt.Errorf("failed to parse data.json: %w", err)
	}
	return &data, nil
}
