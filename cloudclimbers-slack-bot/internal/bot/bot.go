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
	rtm := b.api.NewRTM()
	go rtm.ManageConnection()

	clientOptions := options.Client().ApplyURI(b.config.MongoURI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return err
	}
	db := client.Database("mydatabase")
	pluginRepo := mongodb.NewMongoDBRepository(client, db)

	http.HandleFunc("/interaction", mainplugin.NewMainPlugin(b.config, pluginRepo).ServeHTTP)
	go http.ListenAndServe(":8080", nil)

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			b.eventHandler.HandleMessageEvent(ev)
		}
	}
	return nil
}
