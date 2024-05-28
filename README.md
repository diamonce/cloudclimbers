# Cloud Climbers Slack Bot

![Preview Environments via Slack](images/about.png)



##
## 🌐 **Design Data**
- **[Architecture decisions recod](/ADR)**
- **[High-Level Design](/HLD/high-level-design.md)**
- **[Plugins explained](/plugins-expl)**


##
## Main goal
🚀 **Easily created Preview Environments**: 
- The Cloud Climbers Slack Bot helps software teams increase their development velocity by reducing the time it takes to test and release new features. It allows for the creation, status check, and deletion of preview environments for specific application version directly from Slack. 



##
## Key Features

🧩 **Modern Approach**:
- Utilizes a "best-practices" pull GitOps strategy with Flux, ensuring secure, standardized operations.

🤖 **Artificial Intelligence** 
- Bot has AI plugin augmented with Kubernetes environment preview status and other data to help developer understand possible issues, get useful stats etc.
<!-- Image to illustrate AI
![Image to illustrate AI](images/ai.png) -->

🧩 **Language agnostic**:
- The Cloud Climbers Slack Bot supports community contributions for plugin development in <span style="color:red">any programming language</span>. Whether you prefer Python, Go, JavaScript, or any other language, you can create plugins that interact with the bot through simple HTTP endpoints. Because plugins are Docker containers. 🤓

🛠 **Extensible and Customizable**: 
- The bot's architecture is designed to be extensible and customizable. Community members can develop and share plugins to extend the functionality of the bot, catering to specific needs and workflows. Add AI, cleanup, FLUX, Jenkins in 5 minutes. It is that easy.

🕹️ **Buttons in Slack** 
- Not commands. Because buttons are more robust for user interfaces.

🔐 **Secure** 
- The only external connectivity bot has is websocket connection to Slack servers.






##
### Key Technologies and Protocols

- **Slack Events API & Socket Mode**: To receive real-time events from Slack.
- **Go**: For implementing the bot.
- **HTTP/REST**: For communication between the bot and plugins.
- **YAML**: For configuring plugins and their endpoints.
- **JSON**: For the payloads sent between the bot and plugins.
- **FLUX**: Heart of environment creation.

This setup allows for a highly flexible and extendable bot architecture, encouraging community contributions! 🌍👨‍💻👩‍💻

---

##
### 🍿 Getting Started
- Clone the Repository: Clone the Cloud Climbers Slack Bot repository to your local machine.
- Configure the Bot: Update the YAML configuration file with your Slack tokens, MongoDB URI, and plugin URLs.
- Run the Bot: Use Makefile to build and run the bot and its plugins.
- Develop Plugins: Create your plugins and register them in the YAML configuration file.

---
##
### 🌐 We are the "Cloud Climbers" - Hackathon Team (Credits)

---

#### 🚀 **Members**

- [**Dmytro Chernenko**](https://github.com/diamonce) (Team Lead)
- **Vladyslav Plaksa**
- **Svitlana Dmytrenko**
- [**Denis Klopotovskis**](https://github.com/denisklp)
- [**Andrij Zelenyy**](https://github.com/AZelyony)



#### 🕒 **Work Style**
- **Asynchronous** - One Zoom call per week %)

#### 📊 **Planning Style**
- **Agile** - We plan in Sprints but use 1 Story Point = 1 Hour
- **Team Capacity** 48 H / Sprint. Each team member can do 6 hours in a week.

#### 🎯 **Goal**
- **Minimum Viable Product** in the form of GitHub release.

#### 🛠 **Tools**
- **Project Management:** [JIRA](https://mindocloud.atlassian.net/)
- **Diagrams:** We use Miro boards [Miro](https://miro.com/)
- **Chat:** Telegram Group

##
#### Quick Start
Start your day with setting up the environment

- **Install pre-commit**: `brew install pre-commit`
- **Set up pre-commit hooks**: `pre-commit install`
- **Run pre-commit on all files**: `pre-commit run --all-files`
