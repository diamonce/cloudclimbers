package bot

import (
    "my-slack-bot/config"
    "my-slack-bot/internal/handlers"
    mainplugin "my-slack-bot/internal/plugins/main"
    "my-slack-bot/internal/repository/mongodb"
    "github.com/slack-go/slack"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "net/http"
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
