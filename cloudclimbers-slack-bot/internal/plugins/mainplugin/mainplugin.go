package mainplugin

import (
	"bytes"
	"cloudclimbers-slack-bot/config"
	"cloudclimbers-slack-bot/internal/models"
	"cloudclimbers-slack-bot/internal/repository"
	"cloudclimbers-slack-bot/internal/utils"
	"encoding/json"
	"io"
	"net/http"
	"strings"
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
			p.logAction(action.ActionID, payload.User.ID, payload.Channel.ID)
			switch action.ActionID {
			case "list_enabled_plugins":
				p.ListEnabledPlugins(payload)
			case "help":
				p.PublishHelp(payload)
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
			buttonText := button.Text
			if button.Emoji != "" {
				buttonText = button.Emoji + " " + buttonText
			}
			buttons = append(buttons, slack.NewButtonBlockElement(
				button.ActionID,
				"",
				slack.NewTextBlockObject("plain_text", buttonText, false, false),
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
	// Split actionID to get sub-actions
	actionParts := strings.Split(actionID, "::")
	baseActionID := actionParts[0]

	var pluginConfig *config.PluginConfig
	actionFound := false

	for _, plugin := range p.config.Plugins {
		for _, button := range plugin.Buttons {
			if button.ActionID == baseActionID {
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

	// Create payload with commands, variables, and hash
	data := map[string]interface{}{
		"payload":       payload,
		"commands":      pluginConfig.Commands,
		"variables":     pluginConfig.Variables,
		"hash":          pluginConfig.Hash,
		"sub_action_id": strings.Join(actionParts[1:], "::"),
	}

	payloadBytes, _ := json.Marshal(data)
	resp, err := http.Post(pluginConfig.URL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		utils.Logger().Error("Failed to call plugin", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Logger().Error("Failed to read plugin response", zap.Error(err))
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		utils.Logger().Error("Failed to parse response from plugin", zap.Error(err))
		return
	}

	utils.Logger().Info("Received response from plugin", zap.Any("response", response))

	p.handlePluginResponse(response, payload.Channel.ID)
}

func (p *MainPlugin) handlePluginResponse(response map[string]interface{}, channelID string) {
	utils.Logger().Info("Handling plugin response", zap.Any("response", response))

	var blocks []slack.Block

	if message, ok := response["text"].(string); ok {
		utils.Logger().Info("Plugin response text", zap.String("text", message))
		section := slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", message, false, false), nil, nil)
		blocks = append(blocks, section)
	}

	if blocksFromResponse, ok := response["blocks"].([]interface{}); ok {
		utils.Logger().Info("Processing blocks from plugin response", zap.Any("blocks", blocksFromResponse))
		for _, block := range blocksFromResponse {
			blockMap, ok := block.(map[string]interface{})
			if !ok {
				utils.Logger().Error("Invalid block format", zap.Any("block", block))
				continue
			}

			blockType, _ := blockMap["type"].(string)
			switch blockType {
			case "section":
				text, _ := blockMap["text"].(map[string]interface{})
				textObj := slack.NewTextBlockObject("mrkdwn", text["text"].(string), false, false)
				sectionBlock := slack.NewSectionBlock(textObj, nil, nil)
				blocks = append(blocks, sectionBlock)
			case "input":
				label, _ := blockMap["label"].(map[string]interface{})
				element, _ := blockMap["element"].(map[string]interface{})
				elementType, _ := element["type"].(string)
				var inputElement slack.BlockElement
				switch elementType {
				case "plain_text_input":
					actionID, _ := element["action_id"].(string)
					placeholder, _ := element["placeholder"].(map[string]interface{})
					placeholderObj := slack.NewTextBlockObject("plain_text", placeholder["text"].(string), false, false)
					inputElement = slack.NewPlainTextInputBlockElement(placeholderObj, actionID)
				}
				inputBlock := slack.NewInputBlock(
					blockMap["block_id"].(string),
					slack.NewTextBlockObject("plain_text", label["text"].(string), false, false),
					inputElement,
				)
				blocks = append(blocks, inputBlock)
			}
		}
	}

	if buttons, ok := response["buttons"].([]interface{}); ok {
		utils.Logger().Info("Processing buttons from plugin response", zap.Any("buttons", buttons))
		buttonElements := make([]slack.BlockElement, len(buttons))
		for i, button := range buttons {
			buttonMap, ok := button.(map[string]interface{})
			if !ok {
				utils.Logger().Error("Invalid button format", zap.Any("button", button))
				continue
			}
			text, _ := buttonMap["text"].(string)
			actionID, _ := buttonMap["action_id"].(string)
			emoji, _ := buttonMap["emoji"].(string)
			buttonText := text
			if emoji != "" {
				buttonText = emoji + " " + text
			}
			buttonElements[i] = slack.NewButtonBlockElement(actionID, "", slack.NewTextBlockObject("plain_text", buttonText, false, false))
		}
		actionBlock := slack.NewActionBlock("", buttonElements...)
		blocks = append(blocks, actionBlock)
	}

	if len(blocks) > 0 {
		msg := slack.MsgOptionBlocks(blocks...)
		_, _, err := p.slackClient.PostMessage(channelID, msg)
		if err != nil {
			utils.Logger().Error("Failed to post message", zap.Error(err))
		}
	}
}

func (p *MainPlugin) PublishHelp(payload slack.InteractionCallback) {
	utils.Logger().Info("Publishing help")

	blocks := []slack.Block{
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", "*Help Information*\n\nThis bot allows you to manage preview environments. You can create, get status, and delete environments using the buttons provided.\n\nCommands:\n\n*Create Environment*: Creates a new preview environment.\n*Get Environment Status*: Retrieves the status of the current environment.\n*Delete Environment*: Deletes the current environment.\n\nMake sure you have the correct permissions to perform these actions.", false, false),
			nil,
			nil,
		),
	}

	msg := slack.MsgOptionBlocks(blocks...)

	_, _, err := p.slackClient.PostMessage(payload.Channel.ID, msg)
	if err != nil {
		utils.Logger().Error("Failed to post help message", zap.Error(err))
	}
}

func (p *MainPlugin) PublishHomeTab(userID string) {
	logger := utils.Logger()

	homeView := slack.HomeTabViewRequest{
		Type: "home",
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewImageBlock(
					"https://assets-global.website-files.com/636dbee261df29040c8db281/63785bf38d9dda3badc6329d_VH1jOJ2IcbbIcrMwnC6WVZYbGl7a62ZKCSefB_TumczrstgLmgLfl_wvCtpIEQYEQ6mI4NV4dCo8F7zi4Q3DUeawx4fSmpZtLWIbMV3REqpB-SFhW3boXHLmHK5kH9fu-aCxjcKe_SqJcGe8hCnCqH3qyhK30IjjiMK6wh8W7H-8oYeMWb25VEujkAPsqA.png",
					"Ephemeral preview environment model",
					"image1",
					slack.NewTextBlockObject("plain_text", "Ephemeral preview environment model", false, false),
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

	logger.Info("Publishing home tab", zap.Any("home_view", homeView))

	_, err := p.slackClient.PublishView(userID, homeView, "")
	if err != nil {
		logger.Error("Failed to publish home tab", zap.Error(err))
	}
}

func (p *MainPlugin) logAction(actionID, userID, channelID string) {
	actionLog := &models.ActionLog{
		ActionID:  actionID,
		UserID:    userID,
		ChannelID: channelID,
		Timestamp: time.Now(),
	}

	err := p.pluginRepo.LogAction(actionLog)
	if err != nil {
		utils.Logger().Error("Failed to log action", zap.Error(err))
	}
}

func convertToSlackFields(fields []interface{}) []slack.AttachmentField {
	slackFields := make([]slack.AttachmentField, len(fields))
	for i, field := range fields {
		fieldMap, ok := field.(map[string]interface{})
		if !ok {
			utils.Logger().Error("Invalid field format", zap.Any("field", field))
			continue
		}
		slackFields[i] = slack.AttachmentField{
			Title: fieldMap["title"].(string),
			Value: fieldMap["value"].(string),
			Short: fieldMap["short"].(bool),
		}
	}
	return slackFields
}
