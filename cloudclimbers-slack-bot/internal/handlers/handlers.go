package handlers

import (
    "bytes"
    "encoding/json"
    "log"
    "cloudclimbers-slack-bot/config"
    "net/http"
    "github.com/slack-go/slack"
)

type EventHandler struct {
    plugins map[string]config.PluginConfig
}

func NewEventHandler(plugins map[string]config.PluginConfig) *EventHandler {
    return &EventHandler{
        plugins: plugins,
    }
}

func (h *EventHandler) HandleMessageEvent(ev *slack.MessageEvent) {
    // Assuming the action ID is included in the message text for simplicity
    actionID := ev.Text

    var pluginURL string
    switch actionID {
    case "create_environment":
        pluginURL = h.plugins["create"].URL
    case "get_environment_status":
        pluginURL = h.plugins["get"].URL
    case "delete_environment":
        pluginURL = h.plugins["delete"].URL
    default:
        log.Printf("Unknown action ID: %s", actionID)
        return
    }

    payload, _ := json.Marshal(ev)
    resp, err := http.Post(pluginURL, "application/json", bytes.NewBuffer(payload))
    if err != nil {
        log.Printf("Failed to call plugin: %v", err)
        return
    }
    defer resp.Body.Close()

    // Handle response from plugin
}
