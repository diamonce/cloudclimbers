from flask import Flask, request, jsonify
import logging
import os
import requests

app = Flask(__name__)

# Configure logging
logging.basicConfig(level=logging.INFO)

# Environment variables
ACCOUNT_ID = os.getenv("CLOUDFLARE_ACCOUNT_ID", "your-account-id")
AUTH_TOKEN = os.getenv("CLOUDFLARE_AUTH_TOKEN", "your-auth-token")
CLOUDFLARE_AI_URL = f"https://api.cloudflare.com/client/v4/accounts/{ACCOUNT_ID}/ai/run/@cf/meta/llama-3-8b-instruct"


@app.route("/expert-ai-model", methods=["POST"])
def ask_ai():
    data = request.json

    # Log the incoming request
    logging.info("Received request: %s", data)

    # Extract user inputs from the Slack payload
    payload = data.get("payload", {})
    user_inputs = payload.get("state", {}).get("values", {})

    # If there are no user inputs, prompt the user to enter a question
    if not user_inputs:
        response = {
            "text": "Hello! I'm Cloud Climbers AI, your friendly assistant. Please type in your question below:",
            "blocks": [
                {
                    "type": "input",
                    "block_id": "user_question",
                    "element": {
                        "type": "plain_text_input",
                        "action_id": "question_input",
                        "placeholder": {
                            "type": "plain_text",
                            "text": "Enter your question here",
                        },
                    },
                    "label": {"type": "plain_text", "text": "Question"},
                }
            ],
            "buttons": [
                {
                    "type": "button",
                    "text": "Submit",
                    "action_id": "ask_ai",
                }
            ],
        }
        return jsonify(response)

    # Process the user input
    question = (
        user_inputs.get("user_question", {}).get("question_input", {}).get("value", "")
    )

    # Log the user question
    logging.info("User question: %s", question)

    # Setup Cloudflare AI request
    headers = {
        "Authorization": f"Bearer {AUTH_TOKEN}",
        "Content-Type": "application/json",
    }

    data = {
        "messages": [
            {
                "role": "system",
                "content": "Here is Cloud Climbers AI reply Pods: All core components, including the Slack Bot, MongoDB, and various plugins, are running without issues. Total Pods: 19 (all in Running status except one completed init job).Services: Multiple ClusterIP services are set up for different plugins and the core bot. IP Addresses:cloudclimbers-slack-bot-argocd-redis: 10.74.221.135 cloudclimbers-slack-bot-argocd-repo-server: 10.74.218.195 cloudclimbers-slack-bot-argocd-server: 10.74.219.224 cloudclimbers-slack-bot-mongodb: 10.74.213.236 create-argo-plugin: 10.74.218.133 create-flux-plugin: 10.74.217.186 delete-plugin: 10.74.219.123 expert-ai-model: 10.74.209.235 get-plugin: 10.74.214.88 mongodb: 10.74.222.78 slack-bot: 10.74.220.142 Deployments: Total Deployments: 21 (all ready and up-to-date). Core deployments like ArgoCD, MongoDB, and various plugins are functioning correctly.ReplicaSets: All ReplicaSets are maintaining desired and ready states.StatefulSets: StatefulSets: 2 (all ready).Jobs: A completed init job for the ArgoCD Redis secret indicates setup tasks have finished successfully.Namespaces: Key Namespaces: cloudclimbers: Main namespace for the Slack Bot and plugins. flux-system: Namespace for Flux CD components. cert-manager: Namespace for certificate management. gmp-system: Namespace for Google Managed Prometheus.Key Notes: All components appear to be deployed and running correctly. No critical errors or pending pods were observed. Services are primarily configured for internal communication without external access.: ",
            },
            {"role": "user", "content": question},
        ]
    }

    logging.info("Cloud url: %s", CLOUDFLARE_AI_URL)
    logging.info("Cloud token: %s", AUTH_TOKEN)

    # Make the API request
    response = requests.post(CLOUDFLARE_AI_URL, headers=headers, json=data)
    ai_response = response.json()

    # Log the AI response
    logging.info("AI Response: %s", ai_response)

    # Extract the AI response content
    ai_content = ai_response.get("result", {}).get(
        "response", "Sorry, I couldn't understand that."
    )

    # Log the AI response
    logging.info("AI Content in text: %s", ai_content)

    # Return the AI response to the user in Slack
    response = {
        "text": "Your question was: "
        + question
        + ".\n Here's the response from Cloud Climbers AI on your question: "
        + ai_content,
        "attachments": [],
        "buttons": [
            {
                "type": "button",
                "text": "Ask Another Question",
                "action_id": "ask_ai",
            }
        ],
    }
    return jsonify(response)


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8084)
