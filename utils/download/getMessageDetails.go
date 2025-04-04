package download

import (
	"context"

	"github.com/gotd/td/tg"
)

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
