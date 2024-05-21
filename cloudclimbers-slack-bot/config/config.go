package config

import (
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

type ButtonConfig struct {
	Text     string `mapstructure:"text"`
	ActionID string `mapstructure:"action_id"`
	Emoji    string `mapstructure:"emoji"`
}

type PluginConfig struct {
	URL     string         `mapstructure:"url"`
	Buttons []ButtonConfig `mapstructure:"buttons"`
}

type Config struct {
	SlackBotToken string                  `mapstructure:"slack_bot_token"`
	SlackAppToken string                  `mapstructure:"slack_app_token"`
	MongoURI      string                  `mapstructure:"mongo_uri"`
	DatabaseName  string                  `mapstructure:"database_name"`
	Plugins       map[string]PluginConfig `mapstructure:"plugins"`
	MainButtons   []ButtonConfig          `mapstructure:"main_buttons"`
}

func LoadConfig() (*Config, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	exeDir := filepath.Dir(exePath)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(exeDir)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
