package mainplugin

import (
	"bytes"
	"cloudclimbers-slack-bot/config"
	"cloudclimbers-slack-bot/internal/repository"
	"cloudclimbers-slack-bot/internal/utils"
	"encoding/json"
	"net/http"

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

	// Read plugins from the YAML configuration and check which ones are enabled
	enabledPlugins := []config.PluginConfig{}

	for pluginName, pluginConfig := range p.config.Plugins {
		// Here we assume that a plugin is enabled if it has a valid URL
		if pluginConfig.URL != "" {
			utils.Logger().Info("Enabled plugin found", zap.String("plugin", pluginName))
			enabledPlugins = append(enabledPlugins, pluginConfig)
		}
	}

	// Create buttons for each enabled plugin
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

	// Create a Slack message block with the buttons
	actionBlock := slack.NewActionBlock("enabled_plugins", buttons...)
	msg := slack.MsgOptionBlocks(actionBlock)

	// Send the message to Slack
	_, _, err := p.slackClient.PostMessage(payload.Channel.ID, msg)
	if err != nil {
		utils.Logger().Error("Failed to post message", zap.Error(err))
	}
	utils.Logger().Info("Posted message with enabled plugins")
}

func (p *MainPlugin) ForwardAction(actionID string, payload slack.InteractionCallback) {
	pluginConfig, exists := p.config.Plugins[actionID]
	if !exists {
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
