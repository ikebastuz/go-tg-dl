package download

import (
	"os"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/schollz/progressbar/v3"
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
