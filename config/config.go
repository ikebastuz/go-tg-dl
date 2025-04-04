package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	// API credentials
	ApiId               int    `mapstructure:"api_id"`
	ApiHash             string `mapstructure:"api_hash"`
	Phone               string `mapstructure:"phone"`
	ChannelListSize     int    `mapstructure:"channel_list_size"`
	BatchSize           int    `mapstructure:"batch_size"`
	OldMessageThreshold int    `mapstructure:"old_message_threshold"`
}

// Default values for config
var defaults = Config{
	ApiId:               0,
	ApiHash:             "",
	Phone:               "",
	ChannelListSize:     30,
	BatchSize:           100,
	OldMessageThreshold: 5,
}

// setDefaults configures viper with default values
func setDefaults() {
	viper.SetDefault("api_id", defaults.ApiId)
	viper.SetDefault("api_hash", defaults.ApiHash)
	viper.SetDefault("phone", defaults.Phone)
	viper.SetDefault("channel_list_size", defaults.ChannelListSize)
	viper.SetDefault("batch_size", defaults.BatchSize)
	viper.SetDefault("old_message_threshold", defaults.OldMessageThreshold)
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	setDefaults()

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, create one with defaults
			if err := HandleInit(); err != nil {
				return nil, fmt.Errorf("failed to create default config: %w", err)
			}
			return &defaults, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate required fields
	if cfg.ApiId == 0 || cfg.ApiHash == "" || cfg.Phone == "" {
		return nil, fmt.Errorf("api_id and api_hash and phone must be set in config.json (get them from https://my.telegram.org)")
	}

	return &cfg, nil
}

func HandleInit() error {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	setDefaults()

	// Try to read existing config
	var existingConfig Config
	if err := viper.ReadInConfig(); err == nil {
		// Config exists, unmarshal it
		if err := viper.Unmarshal(&existingConfig); err != nil {
			return fmt.Errorf("failed to parse existing config: %w", err)
		}
		fmt.Println("Updating existing config.json with any new fields...")

		// Explicitly set missing fields from defaults
		if existingConfig.ApiId == 0 {
			viper.Set("api_id", defaults.ApiId)
		}
		if existingConfig.ApiHash == "" {
			viper.Set("api_hash", defaults.ApiHash)
		}
		if existingConfig.Phone == "" {
			viper.Set("phone", defaults.Phone)
		}
		if existingConfig.ChannelListSize == 0 {
			viper.Set("channel_list_size", defaults.ChannelListSize)
		}
		if existingConfig.BatchSize == 0 {
			viper.Set("batch_size", defaults.BatchSize)
		}
		if existingConfig.OldMessageThreshold == 0 {
			viper.Set("old_message_threshold", defaults.OldMessageThreshold)
		}
	} else {
		fmt.Println("Creating new config.json...")
		// For new config, set all values explicitly
		viper.Set("api_id", defaults.ApiId)
		viper.Set("api_hash", defaults.ApiHash)
		viper.Set("phone", defaults.Phone)
		viper.Set("channel_list_size", defaults.ChannelListSize)
		viper.Set("batch_size", defaults.BatchSize)
		viper.Set("old_message_threshold", defaults.OldMessageThreshold)
	}

	// Write config back
	if err := viper.WriteConfig(); err != nil {
		// If file doesn't exist, create it
		if err := viper.SafeWriteConfig(); err != nil {
			return fmt.Errorf("failed to save config file: %w", err)
		}
	}

	fmt.Println("Done! If this is a new config, please edit it and set your api_id and api_hash from https://my.telegram.org")
	return nil
}
