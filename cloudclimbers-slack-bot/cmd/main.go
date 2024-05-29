package main

import (
	"cloudclimbers-slack-bot/config"
	"cloudclimbers-slack-bot/internal/bot"
	"cloudclimbers-slack-bot/internal/handlers"
	"cloudclimbers-slack-bot/internal/utils"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
	// "k8s.io/client-go/kubernetes"
	// "k8s.io/client-go/rest"
)

func main() {
	utils.InitLogger()
	logger := utils.Logger()
	defer logger.Sync()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Initialize Slack API client
	api := slack.New(cfg.SlackBotToken, slack.OptionAppLevelToken(cfg.SlackAppToken))

	eventHandler := handlers.NewEventHandler(api, cfg.Plugins, logger)
	slackBot := bot.NewBot(cfg, eventHandler)

	if err := slackBot.Start(); err != nil {
		logger.Fatal("Failed to start bot", zap.Error(err))
	}
}
