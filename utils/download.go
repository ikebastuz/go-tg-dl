package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"github.com/schollz/progressbar/v3"

	"tg-dl/constants"
	"tg-dl/types"
)

type downloadState struct {
	client     *telegram.Client
	channelID  int64
	accessHash int64
	basePath   string
	dataPath   string
	dl         *downloader.Downloader
	mediaBar   *progressbar.ProgressBar
}

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

// getTempFilename returns a temporary filename for downloads in progress
func getTempFilename(filename string) string {
	return filename + ".download"
}

// isFileDownloaded checks if a file exists at the given path
func isFileDownloaded(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir() && info.Size() > 0
}

func initializeDownloadState(client *telegram.Client, channelID int64, accessHash int64, basePath string, totalMediaFiles int) (*downloadState, error) {
	dataPath := filepath.Join(basePath, "data")
	if err := ensureDir(dataPath); err != nil {
		return nil, fmt.Errorf("error creating directory: %v", err)
	}

	mediaBar := progressbar.NewOptions(totalMediaFiles,
		progressbar.OptionSetDescription("Downloading file ID: 0"),
		progressbar.OptionSetItsString("file"),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	return &downloadState{
		client:     client,
		channelID:  channelID,
		accessHash: accessHash,
		basePath:   basePath,
		dataPath:   dataPath,
		dl:         downloader.NewDownloader(),
		mediaBar:   mediaBar,
	}, nil
}

func (s *downloadState) getMessageDetails(ctx context.Context, msgID int) (*tg.Message, error) {
	message, err := s.client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer: &tg.InputPeerChannel{
			ChannelID:  s.channelID,
			AccessHash: s.accessHash,
		},
		Limit:    1,
		OffsetID: msgID + 1,
	})
	if err != nil {
		return nil, err
	}

	switch m := message.(type) {
	case *tg.MessagesMessages:
		if len(m.Messages) > 0 {
			if msg, ok := m.Messages[0].(*tg.Message); ok {
				return msg, nil
			}
		}
	case *tg.MessagesMessagesSlice:
		if len(m.Messages) > 0 {
			if msg, ok := m.Messages[0].(*tg.Message); ok {
				return msg, nil
			}
		}
	case *tg.MessagesChannelMessages:
		if len(m.Messages) > 0 {
			if msg, ok := m.Messages[0].(*tg.Message); ok {
				return msg, nil
			}
		}
	}
	return nil, nil
}

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
			if wait, ok := HandleFloodWait(err); ok {
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
			if wait, ok := HandleFloodWait(err); ok {
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

func DownloadMedia(ctx context.Context, client *telegram.Client, channelID int64, basePath string) error {
	data, err := loadMessages(basePath)
	if err != nil {
		return fmt.Errorf("error loading messages: %w", err)
	}

	totalMediaFiles := 0
	remainingFiles := 0

	// Ensure data directory exists
	dataPath := filepath.Join(basePath, "data")
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return fmt.Errorf("error creating data directory: %w", err)
	}

	// Pre-scan files and build a map for quick lookup
	existingFiles := make(map[string]bool)
	for _, msg := range data.Messages {
		if msg.IsPhoto || msg.IsVideo {
			totalMediaFiles++
			date, _ := time.Parse(time.RFC3339, msg.Date)
			yearDir := filepath.Join(dataPath, fmt.Sprintf("%d", date.Year()))
			monthDir := filepath.Join(yearDir, fmt.Sprintf("%02d", date.Month()))

			var filename string
			if msg.IsPhoto {
				filename = filepath.Join(monthDir, fmt.Sprintf("%d.%s", msg.ID, constants.PhotoExtension))
			} else if msg.IsVideo {
				filename = filepath.Join(monthDir, fmt.Sprintf("%d.%s", msg.ID, constants.VideoExtension))
			} else {
				continue
			}

			if isFileDownloaded(filename) {
				existingFiles[filename] = true
			} else {
				remainingFiles++
			}
		}
	}

	fmt.Printf("Found %d new media files to download (out of %d total)\n", remainingFiles, totalMediaFiles)

	state, err := initializeDownloadState(client, channelID, data.AccessHash, basePath, totalMediaFiles)
	if err != nil {
		return fmt.Errorf("error initializing download state: %w", err)
	}
	defer state.mediaBar.Close()

	// Set initial progress for already completed files
	if len(existingFiles) > 0 {
		state.mediaBar.Add(len(existingFiles))
	}

	// Process each message that needs downloading
	for _, msg := range data.Messages {
		if !msg.IsPhoto && !msg.IsVideo {
			continue
		}

		date, _ := time.Parse(time.RFC3339, msg.Date)
		yearDir := filepath.Join(dataPath, fmt.Sprintf("%d", date.Year()))
		monthDir := filepath.Join(yearDir, fmt.Sprintf("%02d", date.Month()))

		var filename string
		if msg.IsPhoto {
			filename = filepath.Join(monthDir, fmt.Sprintf("%d.%s", msg.ID, constants.PhotoExtension))
		} else if msg.IsVideo {
			filename = filepath.Join(monthDir, fmt.Sprintf("%d.%s", msg.ID, constants.VideoExtension))
		} else {
			continue
		}

		if existingFiles[filename] {
			continue
		}

		if err := state.processMediaMessage(ctx, msg); err != nil {
			fmt.Printf("\nError processing message %d: %v\n", msg.ID, err)
		}

		time.Sleep(2 * time.Second) // Rate limiting
	}

	fmt.Printf("\nDownloaded media files to %s\n", state.dataPath)
	return nil
}
