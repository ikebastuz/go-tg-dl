package sync

import (
	"time"

	"github.com/gotd/td/tg"

	"tg-dl/types"
)

func processMessages(messages []*tg.Message) []types.MessageData {
	var messageDataList []types.MessageData
	for _, message := range messages {
		isPhoto := false
		isVideo := false

		if message.Media != nil {
			switch media := message.Media.(type) {
			case *tg.MessageMediaPhoto:
				isPhoto = true
			case *tg.MessageMediaDocument:
				if doc, ok := media.Document.(*tg.Document); ok {
					for _, attr := range doc.Attributes {
						if _, ok := attr.(*tg.DocumentAttributeVideo); ok {
							isVideo = true
							break
						}
					}
				}
			}
		}

		msgData := types.MessageData{
			ID:        message.ID,
			Date:      time.Unix(int64(message.Date), 0).Format(time.RFC3339),
			Message:   message.Message,
			Views:     int(message.Views),
			IsPhoto:   isPhoto,
			IsVideo:   isVideo,
			GroupedID: message.GroupedID,
			Author:    message.PostAuthor,
		}

		messageDataList = append(messageDataList, msgData)
	}
	return messageDataList
}
