package types

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// Define context keys
const (
	CtxConfigKey       contextKey = "config"
	CtxDownloadPathKey contextKey = "download_path"
)
