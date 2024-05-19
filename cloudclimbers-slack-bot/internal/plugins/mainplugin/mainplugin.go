package mainplugin

import (
	"bytes"
	"cloudclimbers-slack-bot/config"
	"cloudclimbers-slack-bot/internal/repository"
	"cloudclimbers-slack-bot/internal/utils"
	"encoding/json"
	"net/http"
	"strings"

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
		slackClient:  socketClient.Client, // Correctly reference the client instance
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
				p.listEnabledPlugins(payload)
			default:
				p.forwardAction(action.ActionID, payload)
			}
		}
	}
}

func (p *MainPlugin) listEnabledPlugins(payload slack.InteractionCallback) {
	plugins, err := p.pluginRepo.GetEnabledPlugins()
	if err != nil {
		utils.Logger().Error("Failed to get enabled plugins", zap.Error(err))
		return
	}

	enabledPlugins := make([]string, 0, len(plugins))
	for _, plugin := range plugins {
		enabledPlugins = append(enabledPlugins, plugin.Name)
	}

	attachment := slack.Attachment{
		Text: "Enabled Plugins:\n" + strings.Join(enabledPlugins, "\n"),
	}

	msg := slack.MsgOptionAttachments(attachment)
	_, _, err = p.slackClient.PostMessage(payload.Channel.ID, msg)
	if err != nil {
		utils.Logger().Error("Failed to post message", zap.Error(err))
	}
}

func (p *MainPlugin) forwardAction(actionID string, payload slack.InteractionCallback) {
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

// Helper function to convert attachments to slack fields
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
