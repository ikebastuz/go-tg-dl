package download

import (
	"fmt"
	"path/filepath"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/schollz/progressbar/v3"

	"tg-dl/utils"
)

func initializeDownloadState(client *telegram.Client, channelID int64, accessHash int64, basePath string, totalMediaFiles int) (*downloadState, error) {
	dataPath := filepath.Join(basePath, "data")
	if err := utils.EnsureDir(dataPath); err != nil {
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
