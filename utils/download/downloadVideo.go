package download

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gotd/td/tg"

	"tg-dl/utils"
)

func (s *downloadState) downloadVideo(ctx context.Context, doc *tg.Document, filename string) error {
	tempFilename := getTempFilename(filename)
	f, err := os.Create(tempFilename)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer f.Close()

	for {
		_, err = s.dl.Download(s.client.API(), &tg.InputDocumentFileLocation{
			ID:            doc.ID,
			AccessHash:    doc.AccessHash,
			FileReference: doc.FileReference,
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
			return fmt.Errorf("error downloading video: %w", err)
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
