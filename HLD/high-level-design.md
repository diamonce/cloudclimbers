# High level design of Slack bot for preview environment creation
**High level design**
![hld1](hld1.png)

## Interconnection Between Slack Bot and Plugins 🌐🤖✨

Here's how the magic happens behind the scenes when you interact with our Cloud Climbers Slack Bot: When a user clicks a button or type a command, Slack sends the event to the bot via the Events API and Socket Mode using a WebSocket connection. The bot, implemented using Go, processes the event and determines the appropriate plugin based on the action ID specified in the event payload. The bot then sends an HTTP POST request to the plugin's endpoint, which is specified in a YAML configuration file. The plugin, which can be developed in any language and hosted as a container, receives the request, processes the command using provided variables, and responds with a JSON payload containing text and interactive elements like buttons or input fields. The bot processes this response, formats it into a Slack message, and sends it back to the user in the Slack channel, providing a seamless and interactive experience.

##
## **📜 Love logging** 
 ![Buttons](images/activity_logging.jpeg)