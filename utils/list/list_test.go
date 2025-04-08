package list

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tg-dl/types"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
)

func TestSelectChannel(t *testing.T) {
	channels := []*tg.Channel{
		{ID: 1, Title: "Channel 1"},
		{ID: 2, Title: "Channel 2"},
	}

	tests := []struct {
		name        string
		input       string
		wantChannel *tg.Channel
		wantQuit    bool
	}{
		{
			name:        "Valid selection",
			input:       "1\n",
			wantChannel: channels[0],
			wantQuit:    false,
		},
		{
			name:        "Quit command",
			input:       "q\n",
			wantChannel: nil,
			wantQuit:    true,
		},
		{
			name:        "Invalid number",
			input:       "99\n",
			wantChannel: nil,
			wantQuit:    true,
		},
		{
			name:        "Invalid input",
			input:       "invalid\n",
			wantChannel: nil,
			wantQuit:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			gotChannel, gotQuit := selectChannel(reader, channels)
			assert.Equal(t, tt.wantChannel, gotChannel)
			assert.Equal(t, tt.wantQuit, gotQuit)
		})
	}
}

func TestSaveAndLoadData(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "data.json")
	testData := &types.DataContainer{
		ChannelID:  123,
		AccessHash: 456,
		Messages:   []types.MessageData{},
	}

	// Test saveData
	t.Run("SaveData", func(t *testing.T) {
		err := saveData(testData, testPath)
		assert.NoError(t, err)
		assert.FileExists(t, testPath)

		// Verify file contents
		content, err := os.ReadFile(testPath)
		assert.NoError(t, err)

		var savedData types.DataContainer
		err = json.Unmarshal(content, &savedData)
		assert.NoError(t, err)
		assert.Equal(t, testData.ChannelID, savedData.ChannelID)
		assert.Equal(t, testData.AccessHash, savedData.AccessHash)
	})

	// Test loadData
	t.Run("LoadData", func(t *testing.T) {
		loadedData, err := loadData(testPath)
		assert.NoError(t, err)
		assert.Equal(t, testData.ChannelID, loadedData.ChannelID)
		assert.Equal(t, testData.AccessHash, loadedData.AccessHash)
	})

	// Test loadData with non-existent file
	t.Run("LoadData_NonExistentFile", func(t *testing.T) {
		_, err := loadData(filepath.Join(tmpDir, "nonexistent.json"))
		assert.Error(t, err)
	})
}

func TestConfirmOverwrite(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantResult bool
		wantErr    bool
	}{
		{
			name:       "Confirm overwrite",
			input:      "y\n",
			wantResult: true,
			wantErr:    false,
		},
		{
			name:       "Reject overwrite",
			input:      "n\n",
			wantResult: false,
			wantErr:    false,
		},
		{
			name:       "Invalid input still counts as rejection",
			input:      "invalid\n",
			wantResult: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			result, err := confirmOverwrite(reader, "test/path")

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantResult, result)
		})
	}
}
