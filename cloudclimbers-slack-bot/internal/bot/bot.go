package bot

import (
	"context"
	"net/http"
	"strings"

	"cloudclimbers-slack-bot/config"
	"cloudclimbers-slack-bot/internal/handlers"
	"cloudclimbers-slack-bot/internal/plugins/mainplugin"
	"cloudclimbers-slack-bot/internal/repository/mongodb"
	"cloudclimbers-slack-bot/internal/utils"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"log"
	"os"
)

type Bot struct {
	api          *slack.Client
	socketClient *socketmode.Client
	config       *config.Config
	eventHandler *handlers.EventHandler
}

func NewBot(cfg *config.Config, eventHandler *handlers.EventHandler) *Bot {
	logger := utils.Logger()
	logger.Info("Creating new bot instance")

	// Check if SlackBotToken is empty
	if cfg.SlackBotToken == "" {
		logger.Fatal("Slack bot token is missing")
	}

	// Check if SlackAppToken is empty
	if cfg.SlackAppToken == "" {
		logger.Fatal("Slack app token is missing")
	}

	if !strings.HasPrefix(cfg.SlackAppToken, "xapp-") {
		logger.Fatal("Slack app token must have the prefix 'xapp-'")
	}

	api := slack.New(
		cfg.SlackBotToken,
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
		slack.OptionAppLevelToken(cfg.SlackAppToken),
	)

	socketClient := socketmode.New(
		api,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	return &Bot{
		api:          api,
		socketClient: socketClient,
		config:       cfg,
		eventHandler: eventHandler,
	}
}

func (b *Bot) Start() error {
	logger := utils.Logger()

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

	mainPlugin := mainplugin.NewMainPlugin(b.config, pluginRepo, b.socketClient)

	http.HandleFunc("/interaction", mainPlugin.ServeHTTP)
	logger.Info("Initialized main plugin for Slack interactions")

	go func() {
		for evt := range b.socketClient.Events {
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				logger.Info("Connecting to Slack with Socket Mode...")
			case socketmode.EventTypeConnectionError:
				logger.Error("Connection failed. Retrying later...")
			case socketmode.EventTypeConnected:
				logger.Info("Connected to Slack with Socket Mode.")
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					logger.Warn("Ignored unknown event", zap.Any("event", evt))
					continue
				}

				logger.Info("Event received", zap.Any("event", eventsAPIEvent))

				b.socketClient.Ack(*evt.Request)

				switch eventsAPIEvent.Type {
				case slackevents.CallbackEvent:
					innerEvent := eventsAPIEvent.InnerEvent
					switch ev := innerEvent.Data.(type) {
					case *slackevents.AppMentionEvent:
						_, _, err := b.api.PostMessage(ev.Channel, slack.MsgOptionText("Yes, hello.", false))
						if err != nil {
							logger.Error("Failed to post message", zap.Error(err))
						}
					case *slackevents.MemberJoinedChannelEvent:
						logger.Info("User joined channel", zap.String("user", ev.User), zap.String("channel", ev.Channel))
					}
				default:
					b.socketClient.Debugf("unsupported Events API event received")
				}
			case socketmode.EventTypeInteractive:
				callback, ok := evt.Data.(slack.InteractionCallback)
				if !ok {
					logger.Warn("Ignored unknown interaction", zap.Any("event", evt))
					continue
				}

				logger.Info("Interaction received", zap.Any("callback", callback))

				var payload interface{}

				switch callback.Type {
				case slack.InteractionTypeBlockActions:
					b.socketClient.Debugf("button clicked!")
				case slack.InteractionTypeShortcut:
				case slack.InteractionTypeViewSubmission:
				case slack.InteractionTypeDialogSubmission:
				default:
				}

				b.socketClient.Ack(*evt.Request, payload)
			case socketmode.EventTypeSlashCommand:
				cmd, ok := evt.Data.(slack.SlashCommand)
				if !ok {
					logger.Warn("Ignored unknown slash command", zap.Any("event", evt))
					continue
				}

				logger.Info("Slash command received", zap.Any("command", cmd))

				payload := map[string]interface{}{
					"blocks": []slack.Block{
						slack.NewSectionBlock(
							&slack.TextBlockObject{
								Type: slack.MarkdownType,
								Text: "foo",
							},
							nil,
							slack.NewAccessory(
								slack.NewButtonBlockElement(
									"",
									"somevalue",
									&slack.TextBlockObject{
										Type: slack.PlainTextType,
										Text: "bar",
									},
								),
							),
						),
					},
				}

				b.socketClient.Ack(*evt.Request, payload)
			default:
				logger.Error("Unexpected event type received", zap.Any("event", evt))
			}
		}
	}()

	b.socketClient.Run()
	return nil
}
