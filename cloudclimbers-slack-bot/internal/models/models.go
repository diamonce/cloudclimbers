package models

type Environment struct {
	ID     string `json:"id" bson:"id"`
	Name   string `json:"name" bson:"name"`
	Status string `json:"status" bson:"status"`
}

type PluginConfig struct {
	Name    string `bson:"name"`
	Enabled bool   `bson:"enabled"`
}
