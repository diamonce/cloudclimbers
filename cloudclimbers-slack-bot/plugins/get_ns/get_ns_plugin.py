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

# Function for make requests
def get_resources(url, headers, ca_cert_path):
    response = requests.get(url, headers=headers, verify=ca_cert_path)
    logging.info("responce received: %s", response.json)
    if response.status_code == 200:
        return response.json()
    else:
        print(f"Error getting resources: {response.status_code} - {response.text}")
        return None

@app.route("/get_ns", methods=["POST"])
def get_environment():
    data = request.json
    # Implement the logic for getting the environment status here

    # Log the incoming request
    logging.info("Received request: %s", data)

    # Extract user inputs from the Slack payload
    payload = data.get("payload", {})
    user_inputs = payload.get("state", {}).get("values", {})

    # Downloading a service account
    credentials = service_account.Credentials.from_service_account_file(
        token_path,
        scopes=['https://www.googleapis.com/auth/cloud-platform']
    )

    credentials.refresh(Request())

    # Getting token
    token = credentials.token

    # Build headers
    headers = {
        "Authorization": f"Bearer {token}",
        "Accept": "application/json"
    }
    
    # URLS to get all namespaces
    url = f"{api_server}/api/v1/namespaces"

    # Getting all namespaces
    namespaces = get_resources(url, headers, ca_cert_path)

    if namespaces:
        for ns in namespaces['items']:
            response_text += ns['metadata']['name'] + "\n\n"
    else:
        response_text += f"Can't get namespaces.\n\n"

    response = {
        "text": "Namespaces retrieved successfully!"
        + "namespace "
        + namespace
        + ":\n"
        + str(response_text),
        "attachments": [
            {"text": f"Namespaces in cluster {namespace}:\n{response_text}"}
        ],
        "buttons": [
            {
                "type": "button",
                "text": "Get Environment Status",
                "action_id": "get_environment_status",
            }
        ],
    }
    return jsonify(response)


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8082)