package utils

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func HandleFloodWait(err error) (time.Duration, bool) {
	if err == nil {
		return 0, false
	}

	errStr := err.Error()
	if strings.Contains(errStr, "FLOOD_WAIT") {
		parts := strings.Split(errStr, "FLOOD_WAIT")
		if len(parts) > 1 {
			timeStr := strings.Trim(parts[1], " ()")
			if seconds, err := strconv.Atoi(timeStr); err == nil {
				return time.Duration(seconds) * time.Second, true
			}
		}
		return 60 * time.Second, true
	}
	return 0, false
}
