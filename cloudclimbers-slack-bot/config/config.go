package config

import (
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

type Config struct {
	SlackToken   string
	MongoURI     string
	DatabaseName string
	Plugins      map[string]PluginConfig
}

type PluginConfig struct {
	URL string `mapstructure:"url"`
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
