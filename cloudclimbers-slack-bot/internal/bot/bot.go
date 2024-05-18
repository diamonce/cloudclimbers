package bot

import (
	"context"
	"net/http"

	"cloudclimbers-slack-bot/config"
	"cloudclimbers-slack-bot/internal/handlers"
	"cloudclimbers-slack-bot/internal/plugins/mainplugin"
	"cloudclimbers-slack-bot/internal/repository/mongodb"
	"github.com/slack-go/slack"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"cloudclimbers-slack-bot/internal/utils"
	"go.uber.org/zap"
)

type Bot struct {
	api          *slack.Client
	config       *config.Config
	eventHandler *handlers.EventHandler
}

func NewBot(cfg *config.Config, eventHandler *handlers.EventHandler) *Bot {
	return &Bot{
		api:          slack.New(cfg.SlackToken),
		config:       cfg,
		eventHandler: eventHandler,
	}
}

func (b *Bot) Start() error {
	logger := utils.Logger()

	rtm := b.api.NewRTM()
	go rtm.ManageConnection()

	logger.Info("Connecting to MongoDB...", zap.String("uri", b.config.MongoURI))
	clientOptions := options.Client().ApplyURI(b.config.MongoURI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", zap.Error(err))
		return err
	}
	logger.Info("Successfully connected to MongoDB")

	logger.Info("Using database", zap.String("database", b.config.DatabaseName))
	db := client.Database(b.config.DatabaseName)
	pluginRepo := mongodb.NewMongoDBRepository(client, db)

	http.HandleFunc("/interaction", mainplugin.NewMainPlugin(b.config, pluginRepo).ServeHTTP)
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			b.eventHandler.HandleMessageEvent(ev)
		}
	}
	return nil
}
