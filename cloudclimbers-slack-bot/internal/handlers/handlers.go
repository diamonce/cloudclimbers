package handlers

import (
	"bytes"
	"cloudclimbers-slack-bot/config"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

type EventHandler struct {
	api     *slack.Client
	plugins map[string]config.PluginConfig
	logger  *zap.Logger
}

func NewEventHandler(api *slack.Client, plugins map[string]config.PluginConfig, logger *zap.Logger) *EventHandler {
	return &EventHandler{
		api:     api,
		plugins: plugins,
		logger:  logger,
	}
}

func (h *EventHandler) HandleMessageEvent(ev *slack.MessageEvent) {
	actionID := ev.Text
	actionParts := strings.Split(actionID, "::")
	baseActionID := actionParts[0]

	var pluginConfig *config.PluginConfig
	actionFound := false

	for _, plugin := range h.plugins {
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
		h.logger.Warn("Unknown action ID", zap.String("action_id", actionID))
		return
	}

	// Create payload with commands, variables, and hash
	data := map[string]interface{}{
		"event":         ev,
		"commands":      pluginConfig.Commands,
		"variables":     pluginConfig.Variables,
		"hash":          pluginConfig.Hash,
		"sub_action_id": strings.Join(actionParts[1:], "::"),
	}

	payload, _ := json.Marshal(data)
	resp, err := http.Post(pluginConfig.URL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		h.logger.Error("Failed to call plugin", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("Failed to read plugin response", zap.Error(err))
		return
	}

	h.logger.Info("Plugin response", zap.String("response_body", string(body)))

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		h.logger.Error("Failed to parse plugin response", zap.Error(err))
		return
	}

	messageText, ok := response["text"].(string)
	if !ok {
		h.logger.Warn("Plugin response missing 'text' field")
		return
	}

	// Process blocks if they exist
	var blocks []slack.Block
	if blocksData, ok := response["blocks"].([]interface{}); ok {
		blocks = make([]slack.Block, len(blocksData))
		for i, blockData := range blocksData {
			blockMap, ok := blockData.(map[string]interface{})
			if !ok {
				h.logger.Warn("Invalid block format", zap.Int("index", i))
				continue
			}

			blockType, _ := blockMap["type"].(string)
			h.logger.Info("Processing block", zap.String("block_type", blockType))
			switch blockType {
			case "section":
				text, _ := blockMap["text"].(map[string]interface{})
				textType, _ := text["type"].(string)
				textContent, _ := text["text"].(string)
				blockText := slack.NewTextBlockObject(textType, textContent, false, false)
				section := slack.NewSectionBlock(blockText, nil, nil)
				blocks[i] = section
				h.logger.Info("Created section block", zap.String("text", textContent))
			case "image":
				imageURL, _ := blockMap["image_url"].(string)
				altText, _ := blockMap["alt_text"].(string)
				imageBlock := slack.NewImageBlock(imageURL, altText, "", slack.NewTextBlockObject("plain_text", altText, false, false))
				blocks[i] = imageBlock
				h.logger.Info("Created image block", zap.String("image_url", imageURL))
			case "actions":
				actionElements, ok := blockMap["elements"].([]interface{})
				if !ok {
					h.logger.Warn("Invalid elements format in actions block", zap.Int("index", i))
					continue
				}
				actionBlocks := make([]slack.BlockElement, len(actionElements))
				for j, action := range actionElements {
					actionMap, ok := action.(map[string]interface{})
					if !ok {
						h.logger.Warn("Invalid action format in actions block", zap.Int("index", i), zap.Int("element_index", j))
						continue
					}
					actionText, _ := actionMap["text"].(map[string]interface{})
					actionTextType, _ := actionText["type"].(string)
					actionTextContent, _ := actionText["text"].(string)
					actionTextObject := slack.NewTextBlockObject(actionTextType, actionTextContent, false, false)
					actionID, _ := actionMap["action_id"].(string)
					actionBlocks[j] = slack.NewButtonBlockElement(actionID, "", actionTextObject)
					h.logger.Info("Created button", zap.String("action_id", actionID), zap.String("text", actionTextContent))
				}
				blocks[i] = slack.NewActionBlock("", actionBlocks...)
			default:
				h.logger.Warn("Unknown block type", zap.String("block_type", blockType))
			}
		}
	}

	if blocks == nil {
		// Fallback to a simple message if no blocks are defined
		section := slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", messageText, false, false), nil, nil)
		blocks = []slack.Block{section}
	}

	// Create and send message
	msg := slack.MsgOptionBlocks(blocks...)
	_, _, err = h.api.PostMessage(ev.Channel, msg)
	if err != nil {
		h.logger.Error("Failed to post message", zap.Error(err))
	}
}
