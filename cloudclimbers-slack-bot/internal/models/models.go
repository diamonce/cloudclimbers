package models

import "time"

type Environment struct {
	ID     string `json:"id" bson:"id"`
	Name   string `json:"name" bson:"name"`
	Status string `json:"status" bson:"status"`
}

type PluginConfig struct {
	Name    string `bson:"name"`
	Enabled bool   `bson:"enabled"`
}

type ActionLog struct {
	ActionID  string    `bson:"action_id"`
	UserID    string    `bson:"user_id"`
	ChannelID string    `bson:"channel_id"`
	Timestamp time.Time `bson:"timestamp"`
}
