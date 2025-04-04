package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gotd/td/telegram"
)

// PurgeData removes all download tracking data for a channel
func PurgeData(ctx context.Context, client *telegram.Client, basePath string) error {
	dataPath := filepath.Join(basePath, "data")
	if err := os.RemoveAll(dataPath); err != nil {
		return fmt.Errorf("error removing data directory: %w", err)
	}
	fmt.Printf("Purged data directory at %s\n", dataPath)
	return nil
}
