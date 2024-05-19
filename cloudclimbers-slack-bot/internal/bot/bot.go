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
	mainPlugin   *mainplugin.MainPlugin
}

func NewBot(cfg *config.Config, eventHandler *handlers.EventHandler) *Bot {
	logger := utils.Logger()
	logger.Info("Creating new bot instance")

	if cfg.SlackBotToken == "" {
		logger.Fatal("Slack bot token is missing")
	}

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
		mainPlugin:   mainplugin.NewMainPlugin(cfg, nil, socketClient), // Initialize mainPlugin with nil for repo for now
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
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			logger.Error("Failed to disconnect from MongoDB", zap.Error(err))
		}
	}()
	logger.Info("Successfully connected to MongoDB")

	db := client.Database(b.config.DatabaseName)
	pluginRepo := mongodb.NewMongoDBRepository(client, db)
	b.mainPlugin = mainplugin.NewMainPlugin(b.config, pluginRepo, b.socketClient)

	http.HandleFunc("/interaction", b.mainPlugin.ServeHTTP)
	logger.Info("Initialized main plugin for Slack interactions")

	go b.handleSocketEvents()

	b.socketClient.Run()
	return nil
}

func (b *Bot) handleSocketEvents() {
	logger := utils.Logger()

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
				b.socketClient.Debugf("Unsupported Events API event received")
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
				b.handleBlockActions(callback)
			case slack.InteractionTypeShortcut:
			case slack.InteractionTypeViewSubmission:
			case slack.InteractionTypeDialogSubmission:
			default:
				logger.Warn("Unsupported interaction type", zap.String("type", string(callback.Type)))
			}

			b.socketClient.Ack(*evt.Request, payload)
		case socketmode.EventTypeSlashCommand:
			cmd, ok := evt.Data.(slack.SlashCommand)
			if !ok {
				logger.Warn("Ignored unknown slash command", zap.Any("event", evt))
				continue
			}

			logger.Info("Slash command received", zap.Any("command", cmd))

			blocks := []slack.Block{}
			for _, btn := range b.config.MainButtons {
				sectionBlock := slack.NewSectionBlock(
					&slack.TextBlockObject{
						Type: slack.MarkdownType,
						Text: ":rocket: *" + btn.Text + "*",
					},
					nil,
					slack.NewAccessory(
						slack.NewButtonBlockElement(
							btn.ActionID,
							"",
							&slack.TextBlockObject{
								Type: slack.PlainTextType,
								Text: btn.Text,
							},
						),
					),
				)
				blocks = append(blocks, sectionBlock)
			}

			payload := map[string]interface{}{
				"blocks": blocks,
			}

			b.socketClient.Ack(*evt.Request, payload)
		default:
			logger.Error("Unexpected event type received", zap.Any("event", evt))
		}
	}
}

func (b *Bot) handleBlockActions(callback slack.InteractionCallback) {
	for _, action := range callback.ActionCallback.BlockActions {
		switch action.ActionID {
		case "list_enabled_plugins":
			b.mainPlugin.ListEnabledPlugins(callback)
		default:
			b.mainPlugin.ForwardAction(action.ActionID, callback)
		}
	}
}
