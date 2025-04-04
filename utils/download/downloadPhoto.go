package download

import (
	"context"
	"fmt"
	"os"
	"tg-dl/utils"
	"time"

	"github.com/gotd/td/tg"
)

func (s *downloadState) downloadPhoto(ctx context.Context, photo *tg.Photo, filename string) error {
	tempFilename := getTempFilename(filename)
	var largest *tg.PhotoSize
	for _, size := range photo.Sizes {
		if photoSize, ok := size.(*tg.PhotoSize); ok {
			if largest == nil || photoSize.Size > largest.Size {
				largest = photoSize
			}
		}
	}

	if largest == nil {
		return fmt.Errorf("no suitable photo size found")
	}

	f, err := os.Create(tempFilename)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer f.Close()

	for {
		_, err = s.dl.Download(s.client.API(), &tg.InputPhotoFileLocation{
			ID:            photo.ID,
			AccessHash:    photo.AccessHash,
			FileReference: photo.FileReference,
			ThumbSize:     largest.Type,
		}).WithThreads(2).Stream(ctx, f)

		if err != nil {
			if wait, ok := utils.HandleFloodWait(err); ok {
				f.Close()
				fmt.Printf("\nRate limit hit, waiting for %v...\n", wait)
				time.Sleep(wait)
				f, _ = os.Create(tempFilename)
				continue
			}
			os.Remove(tempFilename)
			return fmt.Errorf("error downloading photo: %w", err)
		}
		break
	}

	f.Close()
	if err := os.Rename(tempFilename, filename); err != nil {
		os.Remove(tempFilename)
		return fmt.Errorf("error moving temp file: %w", err)
	}

	return nil
}
