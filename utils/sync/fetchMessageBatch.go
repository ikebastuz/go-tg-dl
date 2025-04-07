package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"

	"tg-dl/utils"
)

func fetchMessageBatch(ctx context.Context, client *telegram.Client, state *syncState, offsetID int, batchSize int) ([]*tg.Message, error) {
	for {
		messages, err := client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer: &tg.InputPeerChannel{
				ChannelID:  state.data.ChannelID,
				AccessHash: state.data.AccessHash,
			},
			Limit:     batchSize,
			OffsetID:  offsetID,
			AddOffset: 0,
		})

		if err != nil {
			if wait, ok := utils.HandleFloodWait(err); ok {
				fmt.Printf("\nRate limit hit, waiting for %v...\n", wait)
				time.Sleep(wait)
				continue
			}
			return nil, fmt.Errorf("failed to get messages: %w", err)
		}

		channelMsgs, ok := messages.(*tg.MessagesChannelMessages)
		if !ok {
			return nil, nil
		}

		var batch []*tg.Message
		for _, msg := range channelMsgs.Messages {
			if message, ok := msg.(*tg.Message); ok {
				batch = append(batch, message)
			}
		}
		return batch, nil
	}
}
