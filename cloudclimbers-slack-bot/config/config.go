package config

import (
    "github.com/spf13/viper"
)

type Config struct {
    SlackToken string
    Plugins    map[string]PluginConfig
}

type PluginConfig struct {
    URL string `mapstructure:"url"`
}

func LoadConfig() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    viper.SetConfigName("plugins")
    if err := viper.MergeInConfig(); err != nil {
        return nil, err
    }

    if err := viper.UnmarshalKey("plugins", &cfg.Plugins); err != nil {
        return nil, err
    }

    return &cfg, nil
}
