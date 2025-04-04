package download

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gotd/td/tg"

	"tg-dl/constants"
	"tg-dl/types"
)

func (s *downloadState) processMediaMessage(ctx context.Context, msg types.MessageData) error {
	date, err := time.Parse(time.RFC3339, msg.Date)
	if err != nil {
		return fmt.Errorf("error parsing date: %w", err)
	}

	yearDir := filepath.Join(s.dataPath, fmt.Sprintf("%d", date.Year()))
	monthDir := filepath.Join(yearDir, fmt.Sprintf("%02d", date.Month()))

	// Create directories if they don't exist
	if err := os.MkdirAll(monthDir, 0755); err != nil {
		return fmt.Errorf("error creating directory %s: %w", monthDir, err)
	}

	tgMessage, err := s.getMessageDetails(ctx, msg.ID)
	if err != nil {
		return fmt.Errorf("error getting message details: %w", err)
	}

	if tgMessage == nil || tgMessage.Media == nil {
		return nil
	}

	s.mediaBar.Describe(fmt.Sprintf("Downloading file ID: %d", msg.ID))

	var downloadErr error
	switch media := tgMessage.Media.(type) {
	case *tg.MessageMediaPhoto:
		if photo, ok := media.Photo.(*tg.Photo); ok {
			filename := filepath.Join(monthDir, fmt.Sprintf("%d.%s", msg.ID, constants.PhotoExtension))
			downloadErr = s.downloadPhoto(ctx, photo, filename)
		}

	case *tg.MessageMediaDocument:
		if doc, ok := media.Document.(*tg.Document); ok {
			// Only process videos
			isVideo := false
			for _, attr := range doc.Attributes {
				if _, ok := attr.(*tg.DocumentAttributeVideo); ok {
					isVideo = true
					break
				}
			}

			if !isVideo {
				return nil
			}

			filename := filepath.Join(monthDir, fmt.Sprintf("%d.%s", msg.ID, constants.VideoExtension))
			downloadErr = s.downloadVideo(ctx, doc, filename)
		}
	}

	s.mediaBar.Add(1)
	return downloadErr
}
