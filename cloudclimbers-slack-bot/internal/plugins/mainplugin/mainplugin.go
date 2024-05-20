package mainplugin

import (
	"bytes"
	"cloudclimbers-slack-bot/config"
	"cloudclimbers-slack-bot/internal/models"
	"cloudclimbers-slack-bot/internal/repository"
	"cloudclimbers-slack-bot/internal/utils"
	"encoding/json"
	"net/http"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"
)

type MainPlugin struct {
	config       *config.Config
	pluginRepo   repository.PluginRepository
	slackClient  slack.Client
	socketClient *socketmode.Client
}

func NewMainPlugin(cfg *config.Config, repo repository.PluginRepository, socketClient *socketmode.Client) *MainPlugin {
	return &MainPlugin{
		config:       cfg,
		pluginRepo:   repo,
		slackClient:  socketClient.Client,
		socketClient: socketClient,
	}
}

func (p *MainPlugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var payload slack.InteractionCallback
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Failed to parse request", http.StatusBadRequest)
		utils.Logger().Error("Failed to parse request", zap.Error(err))
		return
	}

	if payload.Type == slack.InteractionTypeBlockActions {
		for _, action := range payload.ActionCallback.BlockActions {
			utils.Logger().Info("Received block action", zap.String("action_id", action.ActionID), zap.String("user_id", payload.User.ID), zap.String("channel_id", payload.Channel.ID))
			p.logAction(action.ActionID, payload.User.ID, payload.Channel.ID) // Log the action
			switch action.ActionID {
			case "list_enabled_plugins":
				p.ListEnabledPlugins(payload)
			default:
				p.ForwardAction(action.ActionID, payload)
			}
		}
	}
}

func (p *MainPlugin) ListEnabledPlugins(payload slack.InteractionCallback) {
	utils.Logger().Info("Processing list_enabled_plugins action")

	enabledPlugins := []config.PluginConfig{}

	for pluginName, pluginConfig := range p.config.Plugins {
		if pluginConfig.URL != "" {
			utils.Logger().Info("Enabled plugin found", zap.String("plugin", pluginName))
			enabledPlugins = append(enabledPlugins, pluginConfig)
		}
	}

	buttons := make([]slack.BlockElement, 0, len(enabledPlugins))
	for _, plugin := range enabledPlugins {
		for _, button := range plugin.Buttons {
			buttons = append(buttons, slack.NewButtonBlockElement(
				button.ActionID,
				"",
				slack.NewTextBlockObject("plain_text", button.Text, false, false),
			))
		}
	}

	actionBlock := slack.NewActionBlock("enabled_plugins", buttons...)
	msg := slack.MsgOptionBlocks(actionBlock)

	_, _, err := p.slackClient.PostMessage(payload.Channel.ID, msg)
	if err != nil {
		utils.Logger().Error("Failed to post message", zap.Error(err))
	}
	utils.Logger().Info("Posted message with enabled plugins")
}

func (p *MainPlugin) ForwardAction(actionID string, payload slack.InteractionCallback) {
	var pluginConfig *config.PluginConfig
	actionFound := false

	for _, plugin := range p.config.Plugins {
		for _, button := range plugin.Buttons {
			if button.ActionID == actionID {
				pluginConfig = &plugin
				actionFound = true
				break
			}
		}
		if actionFound {
			break
		}
	}

	if !actionFound || pluginConfig == nil {
		utils.Logger().Warn("Unknown action ID", zap.String("action_id", actionID))
		return
	}

	payloadBytes, _ := json.Marshal(payload)
	resp, err := http.Post(pluginConfig.URL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		utils.Logger().Error("Failed to call plugin", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		utils.Logger().Error("Failed to parse response from plugin", zap.Error(err))
		return
	}

	if message, ok := response["text"]; ok {
		attachment := slack.Attachment{
			Text: message.(string),
		}

		if fields, ok := response["fields"]; ok {
			attachment.Fields = convertToSlackFields(fields.([]map[string]interface{}))
		}

		msg := slack.MsgOptionAttachments(attachment)
		_, _, err := p.slackClient.PostMessage(payload.Channel.ID, msg)
		if err != nil {
			utils.Logger().Error("Failed to post message", zap.Error(err))
		}
	}

	if buttons, ok := response["buttons"]; ok {
		actions := make([]slack.BlockElement, len(buttons.([]map[string]interface{})))
		for i, button := range buttons.([]map[string]interface{}) {
			actions[i] = slack.NewButtonBlockElement(
				button["action_id"].(string),
				"",
				slack.NewTextBlockObject("plain_text", button["text"].(string), false, false),
			)
		}

		actionBlock := slack.NewActionBlock("", actions...)
		msg := slack.MsgOptionBlocks(actionBlock)
		_, _, err := p.slackClient.PostMessage(payload.Channel.ID, msg)
		if err != nil {
			utils.Logger().Error("Failed to post message", zap.Error(err))
		}
	}
}

func (p *MainPlugin) logAction(actionID, userID, channelID string) {
	utils.Logger().Info("Logging action", zap.String("action_id", actionID), zap.String("user_id", userID), zap.String("channel_id", channelID))
	actionLog := &models.ActionLog{
		ActionID:  actionID,
		UserID:    userID,
		ChannelID: channelID,
		Timestamp: time.Now(),
	}

	err := p.pluginRepo.LogAction(actionLog)
	if err != nil {
		utils.Logger().Error("Failed to log action", zap.Error(err))
	} else {
		utils.Logger().Info("Successfully logged action", zap.String("action_id", actionID), zap.String("user_id", userID), zap.String("channel_id", channelID))
	}
}

func convertToSlackFields(fields []map[string]interface{}) []slack.AttachmentField {
	slackFields := make([]slack.AttachmentField, len(fields))
	for i, field := range fields {
		slackFields[i] = slack.AttachmentField{
			Title: field["title"].(string),
			Value: field["value"].(string),
			Short: field["short"].(bool),
		}
	}
	return slackFields
}

func (p *MainPlugin) PublishHomeTab(userID string) {
	logger := utils.Logger()

	homeView := slack.HomeTabViewRequest{
		Type: "home", // Corrected to use VT function for ViewTypeHome
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewImageBlock(
					"https://assets-global.website-files.com/636dbee261df29040c8db281/63785bf38d9dda3badc6329d_VH1jOJ2IcbbIcrMwnC6WVZYbGl7a62ZKCSefB_TumczrstgLmgLfl_wvCtpIEQYEQ6mI4NV4dCo8F7zi4Q3DUeawx4fSmpZtLWIbMV3REqpB-SFhW3boXHLmHK5kH9fu-aCxjcKe_SqJcGe8hCnCqH3qyhK30IjjiMK6wh8W7H-8oYeMWb25VEujkAPsqA.png",
					"Ephemeral preview environment model",
					"image1",
					slack.NewTextBlockObject("plain_text", "Ephemeral preview environment model", false, false), // Added alt text as TextBlockObject
				),
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn", "*What are Preview Environments?*\n\nPreview Environments, also known as Ephemeral Environments, help software teams increase their development velocity by reducing the time it takes to test and release new features.\n\nPreview Environments are created on-demand for testing a specific git branch before it's merged. Unlike persistent Staging or Production Environments, they are intended to be short-lived and single-purpose, existing only as long as needed to test a specific feature or bug fix.\n\nThey help teams standardize best practices for code review, enable faster reviews, rapid feedback, and iterative cycles, ultimately reducing the workload on maintainers and team leaders.", false, false),
					nil,
					nil,
				),
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn", "Preview Environments empower teams to shift their testing process to “pre-merge”, making it easier to find bugs, isolate responsibility, and make appropriate changes. They act as a quality gate allowing features to be thoroughly tested in isolation, facilitating feature parallelization or “shifting left”.", false, false),
					nil,
					nil,
				),
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn", "In summary, Preview Environments fill the gap between local testing and Staging/Production environments. They are designed to test individual features in a production-like environment and have a purpose-driven lifecycle, existing only as long as needed for review. Configure them to create efficiency when deploying code changes.", false, false),
					nil,
					nil,
				),
			},
		},
	}

	// Log the home view payload for debugging
	logger.Info("Publishing home tab", zap.Any("home_view", homeView))

	_, err := p.slackClient.PublishView(userID, homeView, "")
	if err != nil {
		logger.Error("Failed to publish home tab", zap.Error(err))
	}
}
