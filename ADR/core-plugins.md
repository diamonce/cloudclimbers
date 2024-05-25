# Architecture Decision for Cloud Climbers Slack Bot: Core and Plugin Approach
# 21.05.2024

## Context

We are developing the Cloud Climbers Slack Bot to interact with users through Slack. The architecture must support flexibility and reusability. Given our requirements, we have chosen a core and plugin approach to maximize these attributes. 

### Key Components and Rationale:

1. **Slack Events API & Socket Mode**: To receive real-time events from Slack, we use the Events API and Socket Mode. This allows the bot to establish a WebSocket connection, enabling instant event handling and interaction with users.
2. **Go**: We implement the bot using Go for its performance, concurrency support, and efficiency, which are crucial for handling real-time events.
3. **HTTP/REST**: For communication between the bot and plugins, we use HTTP/REST. This standard protocol ensures compatibility and ease of integration between the core bot and various plugins.
4. **YAML**: We use YAML files for configuring plugins and their endpoints. YAML's readability and ease of use make it ideal for configuration management.
5. **JSON**: The payloads sent between the bot and plugins are formatted in JSON. JSON's widespread usage and compatibility with HTTP/REST make it a suitable choice for data exchange.

### Benefits of Core and Plugin Approach:
- **Flexibility**: The core bot can interact with various plugins, each designed for specific tasks, allowing for a modular and adaptable system.
- **Reusability**: Plugins can be reused across different projects, reducing development time and effort.
- **Separation of Concerns**: By separating the core logic from plugin functionality, we can maintain and update each component independently, improving overall maintainability.

## Decision

We will implement the Cloud Climbers Slack Bot using a core and plugin architecture. The core bot, written in Go, will handle real-time events from Slack via the Events API and Socket Mode. Communication between the core bot and plugins will occur through HTTP/REST, with configurations specified in YAML and payloads exchanged in JSON.

## Consequences

**Positive Outcomes:**
- **Modularity and Scalability**: The core and plugin architecture allows us to add or update plugins without modifying the core bot, making the system highly modular and scalable.
- **Improved Maintenance**: The separation of core logic and plugin functionality simplifies maintenance and troubleshooting. Each component can be developed, tested, and deployed independently.
- **Enhanced Flexibility**: Different plugins can be developed in any language and hosted as containers, providing flexibility in choosing the best technology for each task.

**Potential Drawbacks:**
- **Increased Latency**: Communication between the core bot and plugins via HTTP/REST could introduce latency, particularly if plugins are hosted remotely. This needs to be managed to ensure a responsive user experience. But, taking into account that this is preview environment, this should not be considered as serious drawback.
- **Dependency Management**: Ensuring compatibility and proper versioning between the core bot and various plugins could become challenging as the number of plugins increases.

Overall, this architecture decision aligns with our goals of flexibility, reusability, and efficient real-time interaction, enabling us to deliver a robust and scalable Slack bot for our users.