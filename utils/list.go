package utils

import (
	"context"
	"fmt"
	"tg-dl/config"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func ListChannels(ctx context.Context, client *telegram.Client, cfg *config.Config) error {
	dialogs, err := client.API().MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      cfg.ChannelListSize,
	})
	if err != nil {
		return fmt.Errorf("failed to get dialogs: %w", err)
	}

	var channels []*tg.Channel
	switch d := dialogs.(type) {
	case *tg.MessagesDialogs:
	case *tg.MessagesDialogsSlice:
		for _, chat := range d.Chats {
			if channel, ok := chat.(*tg.Channel); ok {
				channels = append(channels, channel)
			}
		}
	}

	fmt.Println("\nAvailable channels:")
	for i := 0; i < len(channels); i++ {
		fmt.Printf("[%d]\tID: %d\t%s\n", i+1, channels[i].ID, channels[i].Title)
	}
	return nil
}
