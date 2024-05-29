# Choosing ArgoCD over Flux for Slack Bot Deployment
# 22.05.2024
## Context

Our organization is operating under a very tight timeline, with only two weeks to develop the Slack bot. Given this constraint, it is essential to select tools and technologies that our team is already familiar with to ensure efficient and timely delivery. 

Our apps use HELM charts, so we can easily use GitOps approach.

We need to decide between ArgoCD and Flux for managing the deployment of preview environment of our apps with help of Slack bot. Both tools are capable of GitOps for Kubernetes, but they have different strengths and weaknesses that are relevant to our specific needs and constraints.

### Considerations:
1. **Ease of Use and Familiarity**: Our team has more experience with ArgoCD, which will reduce the learning curve and development time.
2. **Visual Understanding**: ArgoCD provides a more intuitive visual interface, making it easier for developers to understand the deployment status and troubleshoot issues quickly. While Flux also has a GUI (Flux Capacitor), it is newer and less widely adopted, which could pose challenges.
3. **Repository Management**: Using Flux requires an additional repository to store all parameters, adding complexity to the project (e.g., handling GitHub tokens and managing multiple repositories for each preview environment). 
ArgoCD does not require this, because it already have very useful **"argocd app create"** with possibility to pass needed parameters, simplifying the setup and management process.

## Decision

We will use ArgoCD for the continuous deployment deployment of preview environment of our apps with help of Slack bot.

## Consequences

**Positive Outcomes:**
- **Faster Development**: Leveraging the team's existing knowledge of ArgoCD will save time on training and troubleshooting, allowing us to meet our tight development deadline.
- **Simplified Management**: ArgoCD's **"argocd app create"** will allow us to spend less time developing needed functionality.
- **Improved Troubleshooting**: The user-friendly visual interface of ArgoCD will help developers quickly identify and resolve issues, improving overall productivity and reducing downtime.

**Potential Drawbacks:**
- **Missed Opportunities with Flux**: By not using Flux, we might miss out clearest GitOps approach, with separate repos.
- **Dependency on ArgoCD**: Our decision to use ArgoCD means we are reliant on its ecosystem, which could pose challenges if there are any changes or issues with the tool in the future.

Overall, this decision aligns with our current business priorities and technical constraints, ensuring that we can deliver the Slack bot on time while maintaining a manageable level of complexity.

**In future, Flux should be analyzed again, as Flux offers security benefits, as no commands from bot are runned on Kubernetes cluster**