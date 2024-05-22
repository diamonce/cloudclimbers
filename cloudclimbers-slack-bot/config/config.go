package config

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

type Config struct {
	SlackBotToken string                  `mapstructure:"slack_bot_token"`
	SlackAppToken string                  `mapstructure:"slack_app_token"`
	MongoURI      string                  `mapstructure:"mongo_uri"`
	DatabaseName  string                  `mapstructure:"database_name"`
	Plugins       map[string]PluginConfig `mapstructure:"plugins"`
	MainButtons   []MainButton            `mapstructure:"main_buttons"`
}

type MainButton struct {
	Text     string `mapstructure:"text"`
	ActionID string `mapstructure:"action_id"`
	Emoji    string `mapstructure:"emoji"`
}

type PluginConfig struct {
	URL       string            `mapstructure:"url"`
	Buttons   []ButtonConfig    `mapstructure:"buttons"`
	Commands  string            `mapstructure:"commands"`
	Variables map[string]string `mapstructure:"variables"`
	Hash      HashConfig        `mapstructure:"hash"`
}

type ButtonConfig struct {
	Text     string `mapstructure:"text"`
	ActionID string `mapstructure:"action_id"`
	Emoji    string `mapstructure:"emoji"`
}

type HashConfig struct {
	Type  string `mapstructure:"type"`
	Value string `mapstructure:"value"`
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

	// Calculate hash for each plugin's commands and set it in the config
	for pluginName, pluginConfig := range cfg.Plugins {
		hash := calculateHash(pluginConfig.Commands)
		cfg.Plugins[pluginName] = PluginConfig{
			URL:       pluginConfig.URL,
			Buttons:   pluginConfig.Buttons,
			Commands:  pluginConfig.Commands,
			Variables: pluginConfig.Variables,
			Hash: HashConfig{
				Type:  pluginConfig.Hash.Type,
				Value: hash,
			},
		}
	}

	return &cfg, nil
}

func calculateHash(commands string) string {
	hasher := sha256.New()
	hasher.Write([]byte(commands))
	return hex.EncodeToString(hasher.Sum(nil))
}
