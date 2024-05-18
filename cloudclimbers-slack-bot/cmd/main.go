package main

import (
    "log"
    "cloudclimbers-slack-bot/config"
    "cloudclimbers-slack-bot/internal/bot"
    "cloudclimbers-slack-bot/internal/handlers"
    "cloudclimbers-slack-bot/internal/utils"
    "go.uber.org/zap"
)

func main() {
    utils.InitLogger()
    defer utils.Logger().Sync()

    cfg, err := config.LoadConfig()
    if err != nil {
        utils.Logger().Fatal("Could not load config", zap.Error(err))
    }

    eventHandler := handlers.NewEventHandler(cfg.Plugins)
    slackBot := bot.NewBot(cfg, eventHandler)

    if err := slackBot.Start(); err != nil {
        utils.Logger().Fatal("Could not start the bot", zap.Error(err))
    }
}
