from google.oauth2 import service_account
from google.auth.transport.requests import Request
import requests
import logging
from flask import Flask, request, jsonify

app = Flask(__name__)

# Configure logging
logging.basicConfig(level=logging.INFO)

# Connection params
token_path = "/var/secrets/decrypted/service-account-key.json"
namespace = "cloudclimbers"
api_server = "https://kubernetes.default.svc"
# for production need to set valid cert
# ca_cert_path = "/path/to/ca.crt"
ca_cert_path = False


# Read token from file
# def read_token(token_path):
#    with open(token_path, "r") as token_file:
#        return token_file.read().strip()


# Function for make requests
def get_resources(url, headers, ca_cert_path):
    response = requests.get(url, headers=headers, verify=ca_cert_path)
    logging.info("responce received: %s", response.json)
    if response.status_code == 200:
        return response.json()
    else:
        print(f"Error getting resources: {response.status_code} - {response.text}")
        return None


@app.route("/get", methods=["POST"])
def get_environment():
    data = request.json
    # Implement the logic for getting the environment status here

    # Log the incoming request
    logging.info("Received request: %s", data)

    commands = data.get("commands", "")
    variables = data.get("variables", {})
    hash_value = data.get("hash", {})

    # Log the fields, variables, commands, and hash
    logging.info("Commands: %s", commands)
    logging.info("Variables: %s", variables)
    logging.info("Hash: %s", hash_value)

    # Ensure variables is a dictionary even if it's None
    if variables is None:
        variables = {}

    # Extract user inputs from the Slack payload
    payload = data.get("payload", {})
    user_inputs = payload.get("state", {}).get("values", {})

    # Log the user inputs
    logging.info("User inputs: %s", user_inputs)

    # Update variables with user inputs
    for block_id, block_value in user_inputs.items():
        action_id = list(block_value.keys())[0]
        variables[block_id] = block_value[action_id].get("value", "")

    # Log the updated variables
    logging.info("Updated Variables: %s", variables)

    # Check if variables are still missing and need to be provided by the user
    missing_variables = {key: "" for key, value in variables.items() if value == ""}

    if missing_variables:
        # Respond with a prompt for the user to enter missing variables
        input_blocks = []
        for var_name in missing_variables.keys():
            input_blocks.append(
                {
                    "type": "input",
                    "block_id": var_name,
                    "element": {
                        "type": "plain_text_input",
                        "action_id": var_name,
                        "placeholder": {
                            "type": "plain_text",
                            "text": f"Enter {var_name}",
                        },
                    },
                    "label": {"type": "plain_text", "text": f"{var_name}"},
                }
            )

        response = {
            "text": "Please provide the following variables:",
            "blocks": input_blocks,
            "buttons": [
                {
                    "type": "button",
                    "text": "Submit Variables",
                    "action_id": "get_environment_status",
                }
            ],
        }
        return jsonify(response)

    # Downloading a service account
    credentials = service_account.Credentials.from_service_account_file(
        token_path, scopes=["https://www.googleapis.com/auth/cloud-platform"]
    )

    credentials.refresh(Request())

    # Getting token
    token = credentials.token

    # Build headers
    headers = {"Authorization": f"Bearer {token}", "Accept": "application/json"}

    # URLS to get all pods, replicasets, services and other resources in namespace
    urls = {
        "pods": f"{api_server}/api/v1/namespaces/{namespace}/pods",
        "replicasets": f"{api_server}/apis/apps/v1/namespaces/{namespace}/replicasets",
        "deployments": f"{api_server}/apis/apps/v1/namespaces/{namespace}/deployments",
        "services": f"{api_server}/api/v1/namespaces/{namespace}/services",
    }

    # Getting all resources
    resources = {
        resource: get_resources(url, headers, ca_cert_path)
        for resource, url in urls.items()
    }

    # Making answer
    response_text = ""
    for resource, data in resources.items():
        if data:
            resource_names = [item["metadata"]["name"] for item in data["items"]]
            response_text += (
                f"{resource.capitalize()}:\n" + "\n".join(resource_names) + "\n\n"
            )
        else:
            response_text += f"Can't get {resource}.\n\n"

    response = {
        "text": "Environment status retrieved successfully!"
        + "namespace "
        + namespace
        + ":\n"
        + str(response_text),
        "attachments": [
            {"text": f"Resources in namespace {namespace}:\n{response_text}"}
        ],
        "buttons": [
            {
                "type": "button",
                "text": "Delete Environment",
                "action_id": "delete_environment",
            }
        ],
    }
    return jsonify(response)


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8082)
