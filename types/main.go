package types

type DataContainer struct {
	ChannelID  int64         `json:"channelId"`
	AccessHash int64         `json:"accessHash"`
	Messages   []MessageData `json:"messages"`
}

type MessageData struct {
	ID        int    `json:"id"`
	Date      string `json:"date"`
	Message   string `json:"message"`
	Views     int    `json:"views"`
	IsPhoto   bool   `json:"isPhoto"`
	IsVideo   bool   `json:"isVideo"`
	Author    string `json:"author"`
	GroupedID int64  `json:"groupedId,omitempty"`
	FileExt   string `json:"fileExt,omitempty"`
}
