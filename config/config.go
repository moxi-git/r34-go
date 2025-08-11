package config

import (
	"github.com/spf13/viper"
)

// Settings represents the application configuration
type Settings struct {
	Limit  uint16 `mapstructure:"limit"`
	Images bool   `mapstructure:"images"`
	Gif    bool   `mapstructure:"gif"`
	Video  bool   `mapstructure:"video"`
	IsAPI  bool   `mapstructure:"is_api"`
}

var AppSettings Settings

// Init initializes the configuration with default values
func Init() {
	// Set default values
	viper.SetDefault("limit", 100)
	viper.SetDefault("images", true)
	viper.SetDefault("gif", true)
	viper.SetDefault("video", true)
	viper.SetDefault("is_api", true)

	// Set config file properties
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.r34downloader")

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		// If config file doesn't exist, create one with defaults
		viper.SafeWriteConfig()
	}

	// Unmarshal config into AppSettings
	viper.Unmarshal(&AppSettings)
}

// Save saves the current settings to config file
func Save() error {
	viper.Set("limit", AppSettings.Limit)
	viper.Set("images", AppSettings.Images)
	viper.Set("gif", AppSettings.Gif)
	viper.Set("video", AppSettings.Video)
	viper.Set("is_api", AppSettings.IsAPI)
	return viper.WriteConfig()
}
