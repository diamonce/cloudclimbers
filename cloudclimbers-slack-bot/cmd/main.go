package main

import (
	"cloudclimbers-slack-bot/config"
	"cloudclimbers-slack-bot/internal/bot"
	"cloudclimbers-slack-bot/internal/handlers"
	"cloudclimbers-slack-bot/internal/utils"
	"go.uber.org/zap"
)

func main() {
	utils.InitLogger()
	logger := utils.Logger()
	defer logger.Sync()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	eventHandler := handlers.NewEventHandler(cfg.Plugins)
	slackBot := bot.NewBot(cfg, eventHandler)

	if err := slackBot.Start(); err != nil {
		logger.Fatal("Failed to start bot", zap.Error(err))
	}
}
