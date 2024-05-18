package main

import (
    "bytes"
    "encoding/json"
    "log"
    "cloudclimbers-slack-bot/config"
    "cloudclimbers-slack-bot/internal/repository"
    "net/http"
    "strings"

    "github.com/slack-go/slack"
)

type MainPlugin struct {
    config      *config.Config
    pluginRepo  repository.PluginRepository
    slackClient *slack.Client
}

func NewMainPlugin(cfg *config.Config, repo repository.PluginRepository) *MainPlugin {
    return &MainPlugin{
        config:      cfg,
        pluginRepo:  repo,
        slackClient: slack.New(cfg.SlackToken),
    }
}

func (p *MainPlugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    var payload slack.InteractionCallback
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "Failed to parse request", http.StatusBadRequest)
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
        log.Printf("Failed to get enabled plugins: %v", err)
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
        log.Printf("Failed to post message: %v", err)
    }
}

func (p *MainPlugin) forwardAction(actionID string, payload slack.InteractionCallback) {
    pluginConfig, exists := p.config.Plugins[actionID]
    if !exists {
        log.Printf("Unknown action ID: %s", actionID)
        return
    }

    payloadBytes, _ := json.Marshal(payload)
    resp, err := http.Post(pluginConfig.URL, "application/json", bytes.NewBuffer(payloadBytes))
    if err != nil {
        log.Printf("Failed to call plugin: %v", err)
        return
    }
    defer resp.Body.Close()

    var response map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        log.Printf("Failed to parse response from plugin: %v", err)
        return
    }

    if message, ok := response["text"]; ok {
        attachment := slack.Attachment{
            Text: message.(string),
        }

        if attachments, ok := response["attachments"]; ok {
            attachment.Attachments = attachments.([]slack.Attachment)
        }

        msg := slack.MsgOptionAttachments(attachment)
        _, _, err := p.slackClient.PostMessage(payload.Channel.ID, msg)
        if err != nil {
            log.Printf("Failed to post message: %v", err)
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

        section := slack.NewSectionBlock(nil, nil, slack.NewActionBlock("", actions...))
        msg := slack.MsgOptionBlocks(section)
        _, _, err := p.slackClient.PostMessage(payload.Channel.ID, msg)
        if err != nil {
            log.Printf("Failed to post message: %v", err)
        }
    }
}
