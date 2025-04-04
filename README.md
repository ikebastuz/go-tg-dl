# Telegram Channel Downloader

A Go program that downloads messages and media files from Telegram channels. The program uses the `gotd/td` library to interact with Telegram's API.

## Features

- Lists available channels
- Downloads all messages from a selected channel
- Saves messages to a JSON file (`data.json`)
- Downloads photos and videos, organizing them by month
- Shows progress bars for media downloads
- Handles rate limiting and FLOOD_WAIT errors gracefully

## Prerequisites

1. Go 1.16 or later
2. Telegram API credentials (APP_ID and APP_HASH)

To obtain your API credentials:

1. Visit https://my.telegram.org/apps
2. Log in with your phone number
3. Create a new application
4. Note down the `api_id` and `api_hash`

## Installation

```bash
git clone https://github.com/yourusername/tg-dl
cd tg-dl
go mod download
```

## Configuration

Set your Telegram API credentials as environment variables:

```bash
export APP_ID="your_app_id"
export APP_HASH="your_app_hash"
```

## Usage

1. Run the program:

```bash
go run .
```

2. The program will:
   - Connect to Telegram
   - List available channels
   - Prompt you to select a channel
   - Download all messages and save them to `data.json`
   - Download media files to `downloads/<channel_name>/<YYYY-MM>/`

## Output Structure

```
.
├── data.json              # All messages in JSON format
└── downloads/
    └── channel_name/
        ├── 2024-01/      # Media files organized by month
        │   ├── 123.jpg
        │   └── 456.mp4
        └── 2024-02/
            ├── 789.jpg
            └── 012.mp4
```

## Rate Limiting

The program implements rate limiting to handle Telegram's FLOOD_WAIT errors:

- Delays between message batch fetches
- Delays between media downloads
- Automatic retry with appropriate wait times when hitting rate limits

## License

MIT License
