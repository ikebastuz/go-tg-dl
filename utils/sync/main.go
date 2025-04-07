package sync

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/schollz/progressbar/v3"

	"tg-dl/config"
	"tg-dl/types"
)

type syncState struct {
	data      types.DataContainer
	latestID  int
	channelID int64
	basePath  string
	force     bool
}

func SyncMessages(ctx context.Context, client *telegram.Client, basePath string, force bool, cfg *config.Config) error {
	state, err := loadExistingMessages(basePath, force)
	if err != nil {
		return err
	}

	batchSize := cfg.BatchSize
	oldMessageThreshold := cfg.OldMessageThreshold

	var (
		newMessages      []*tg.Message
		offsetID         = 0
		oldMessageCount  = 0
		foundNewMessages = false
	)

	fmt.Println("\nFetching messages...")
	bar := progressbar.Default(-1, "Fetching messages")
	defer bar.Close()

	for {
		batch, err := fetchMessageBatch(ctx, client, state, offsetID, batchSize)
		if err != nil {
			bar.Clear()
			return err
		}

		if len(batch) == 0 {
			break
		}

		foundNewInBatch := false
		var newBatch []*tg.Message
		for _, message := range batch {
			if !force && message.ID <= state.latestID {
				oldMessageCount++
			} else {
				foundNewInBatch = true
				foundNewMessages = true
				newBatch = append(newBatch, message)
			}
		}

		bar.Add(len(newBatch))

		if !force && !foundNewInBatch && oldMessageCount >= oldMessageThreshold {
			if !foundNewMessages {
				bar.Clear()
				fmt.Print("\033[2K") // Clear the current line
				fmt.Println("\nNo new messages found")
				return nil
			}
			break
		}

		if foundNewInBatch {
			oldMessageCount = 0
		}

		newMessages = append(newMessages, newBatch...)
		offsetID = batch[len(batch)-1].ID
		time.Sleep(1 * time.Second) // TODO: to config
	}

	bar.Clear()
	fmt.Print("\033[2K") // Clear the current line

	if len(newMessages) == 0 && !force {
		fmt.Println("\nNo new messages found")
		return nil
	}

	// Sort new messages by date (oldest first)
	sort.Slice(newMessages, func(i, j int) bool {
		return newMessages[i].Date < newMessages[j].Date
	})

	newMessageDataList := processMessages(newMessages)

	var allMessages []types.MessageData
	if force {
		allMessages = newMessageDataList
	} else {
		allMessages = append(state.data.Messages, newMessageDataList...)
	}

	// Sort all messages by ID
	sort.Slice(allMessages, func(i, j int) bool {
		return allMessages[i].ID < allMessages[j].ID
	})

	// Create new data container preserving channel ID
	newData := types.DataContainer{
		ChannelID:  state.data.ChannelID,
		AccessHash: state.data.AccessHash,
		Messages:   allMessages,
	}

	if err := saveMessages(newData, basePath); err != nil {
		return err
	}

	if force {
		fmt.Printf("Saved data.json with %d total messages\n", len(allMessages))
	} else {
		fmt.Printf("Updated data.json with %d total messages (%d new)\n",
			len(allMessages),
			len(newMessageDataList))
	}
	return nil
}
