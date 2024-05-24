# Cloud Climbers Slack Bot

## Key Features

🚀 **Preview Environments**: The Cloud Climbers Slack Bot helps software teams increase their development velocity by reducing the time it takes to test and release new features. It allows for the creation, status check, and deletion of preview environments directly from Slack. And much more...

🧩 **Plugin Development**: The Cloud Climbers Slack Bot supports community contributions for plugin development in <span style="color:red">any programming language</span>. Whether you prefer Python, Go, JavaScript, or any other language, you can create plugins that interact with the bot through simple HTTP endpoints. Because plugins are Docker containers. 🤓

🛠 **Extensible and Customizable**: The bot's architecture is designed to be extensible and customizable. Community members can develop and share plugins to extend the functionality of the bot, catering to specific needs and workflows. Add AI, cleanup, FLUX, Jenkins in 5 minutes. It is that easy.

🕹️ **Buttons in Slack** Not commands. Because buttons are more robust for user interfaces.

🔐 **Secure** The only external connectivity bot has is websocket connection to Slack servers. 

🏗️ **Visual** We use ArgoCD to give visual representation on  what is happening in cluster.

## Interconnection Between Slack Bot and Plugins

When a user interacts with the Cloud Climbers Slack Bot by clicking a button or typing a command, Slack sends the event to the bot via the Events API and Socket Mode using a WebSocket connection. The bot, implemented using Go, processes the event and determines the appropriate plugin based on the action ID specified in the event payload. The bot then sends an HTTP POST request to the plugin's endpoint, which is specified in a YAML configuration file. The plugin, which can be developed in any language and hosted as a container, receives the request, processes the command using provided variables, and responds with a JSON payload containing text and interactive elements like buttons or input fields. The bot processes this response, formats it into a Slack message, and sends it back to the user in the Slack channel, providing a seamless and interactive experience.

### Example Plugin Configuration

```yaml
plugins:
  create:
    url: "http://localhost:8081/create"
    buttons:
      - text: "Create Environment"
        action_id: "create_environment"
  get:
    url: "http://localhost:8082/get"
    buttons:
      - text: "Get Environment Status"
        action_id: "get_environment_status"
  delete:
    url: "http://localhost:8083/delete"
    buttons:
      - text: "Delete Environment"
        action_id: "delete_environment"

main_buttons:
  - text: "List Enabled Plugins"
    action_id: "list_enabled_plugins"
    emoji: ":rocket:"
  - text: "Help"
    action_id: "help"
    emoji: ":information_source:"
```

### 🍿 Getting Started
Clone the Repository: Clone the Cloud Climbers Slack Bot repository to your local machine.
Configure the Bot: Update the YAML configuration file with your Slack tokens, MongoDB URI, and plugin URLs.
Run the Bot: Use Makefile to build and run the bot and its plugins.
Develop Plugins: Create your plugins in any programming language and register them in the YAML configuration file.

---

### 🌐 We are the "Cloud Climbers" - Hackathon Team

---

#### 🚀 **Members**

- **Dmytro Chernenko** (Team Lead)
- **Vladyslav Plaksa**
- **Svitlana Dmytrenko**
- **Denis Klopotovskis**
- **Andrij Zelenyy**

#### 🤖 **Product Focus**
- **Artificial Intelligence (AI)**

#### 🕒 **Work Style**
- **Asynchronous** - One Zoom call per week %)

#### 📊 **Planning Style**
- **Agile** - We plan in Sprints but use 1 Story Point = 1 Hour
- **Team Capacity** 48 H / Sprint. Each team member can do 6 hours in a week.

#### 🎯 **Goal**
- **Minimum Viable Product** in the form of GitHub release.

#### 🛠 **Tools**
- **Project Management:** [JIRA](https://mindocloud.atlassian.net/)
- **CI/CD Diagrams:** We use Miro boards [Miro](https://miro.com/)
- **Chat:** Telegram Group

#### Quick Start
Start your day with setting up the environment

# brew install pre-commit
# pre-commit install
# pre-commit run --all-files
