from google.oauth2 import service_account
from google.auth.transport.requests import Request
import requests
import logging
from datetime import datetime
from flask import Flask, request, jsonify
import json

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


def format_age(timestamp):
    if not timestamp:
        return "N/A"
    timestamp = datetime.strptime(timestamp, "%Y-%m-%dT%H:%M:%SZ")
    age = datetime.utcnow() - timestamp
    days = age.days
    hours, remainder = divmod(age.seconds, 3600)
    minutes, _ = divmod(remainder, 60)
    return f"{days}d{hours}h{minutes}m"


# Function for making requests
def get_resources(url, headers, ca_cert_path):
    response = requests.get(url, headers=headers, verify=ca_cert_path)
    logging.info("Response received from URL %s: %s", url, response.text)
    if response.status_code == 200:
        return response.json()
    else:
        logging.error(
            f"Error getting resources: {response.status_code} - {response.text}"
        )
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

    # Set namespace by user input
    namespace = variables["namespace"]

    # Downloading a service account
    credentials = service_account.Credentials.from_service_account_file(
        token_path, scopes=["https://www.googleapis.com/auth/cloud-platform"]
    )

    credentials.refresh(Request())

    # Getting token
    token = credentials.token

    # Build headers
    headers = {"Authorization": f"Bearer {token}", "Accept": "application/json"}

    # URLs to get all pods, replicasets, services, and other resources in namespace
    urls = {
        #       "pods": f"{api_server}/api/v1/namespaces/{namespace}/pods",
        #       "replicasets": f"{api_server}/apis/apps/v1/namespaces/{namespace}/replicasets",
        "deployments": f"{api_server}/apis/apps/v1/namespaces/{namespace}/deployments",
        "services": f"{api_server}/api/v1/namespaces/{namespace}/services",
    }

    # Getting all resources
    resources = {}
    for resource, url in urls.items():
        logging.info("Requesting %s from URL %s", resource, url)
        resource_data = get_resources(url, headers, ca_cert_path)
        resources[resource] = resource_data
        logging.info("Received data for %s: %s", resource, resource_data)

    response_text = (
        "Environment status retrieved successfully for " + str(namespace) + "! \n```"
    )

    url_ports_list = []

    for resource, data in resources.items():
        if data:
            response_text += f"{resource.capitalize()}:\n"
            if resource == "pods":
                response_text += f"{'NAME':<20}{'READY':<10}{'STATUS':<15}{'RESTARTS':<10}{'AGE':<10}\n"
                for item in data["items"]:
                    name = item["metadata"]["name"]
                    ready = f"{sum(1 for c in item['status']['containerStatuses'] if c['ready'])}/{len(item['status']['containerStatuses'])}"
                    status = item["status"].get("phase", "N/A")
                    restarts = sum(
                        c["restartCount"] for c in item["status"]["containerStatuses"]
                    )
                    age = format_age(item["metadata"]["creationTimestamp"])
                    response_text += (
                        f"{name:<20}{ready:<10}{status:<15}{restarts:<10}{age:<10}\n"
                    )

            elif resource == "replicasets":
                response_text += f"{'NAME':<20}{'DESIRED':<10}{'CURRENT':<10}{'READY':<10}{'AGE':<10}\n"
                for item in data["items"]:
                    name = item["metadata"]["name"]
                    desired = item["spec"].get("replicas", "N/A")
                    current = item["status"].get("replicas", "N/A")
                    ready = item["status"].get("readyReplicas", "N/A")
                    age = format_age(item["metadata"]["creationTimestamp"])
                    response_text += (
                        f"{name:<20}{desired:<10}{current:<10}{ready:<10}{age:<10}\n"
                    )

            elif resource == "deployments":
                response_text += f"{'NAME':<20}{'READY':<10}{'UP-TO-DATE':<10}{'AVAILABLE':<10}{'AGE':<10}\n"
                for item in data["items"]:
                    name = item["metadata"]["name"]
                    ready = f"{item['status'].get('readyReplicas', 0)}/{item['spec'].get('replicas', 0)}"
                    up_to_date = item["status"].get("updatedReplicas", "N/A")
                    available = item["status"].get("availableReplicas", "N/A")
                    age = format_age(item["metadata"]["creationTimestamp"])
                    response_text += f"{name:<20}{ready:<10}{up_to_date:<10}{available:<10}{age:<10}\n"

            elif resource == "services":
                response_text += f"{'NAME':<20}{'TYPE':<15}{'CLUSTER-IP':<20}{'EXTERNAL-IP':<20}{'PORT(S)':<15}{'AGE':<10}\n"
                for item in data["items"]:
                    name = item["metadata"]["name"]
                    svc_type = item["spec"].get("type", "N/A")
                    cluster_ip = item["spec"].get("clusterIP", "N/A")
                    external_ips = [
                        ingress.get("ip", "<none>")
                        for ingress in item["status"]
                        .get("loadBalancer", {})
                        .get("ingress", [])
                    ]
                    external_ip = ", ".join(external_ips) if external_ips else "<none>"
                    ports = ", ".join(
                        [
                            f"{p['port']}/{p['protocol']}"
                            for p in item["spec"].get("ports", [])
                        ]
                    )
                    age = format_age(item["metadata"]["creationTimestamp"])

                    if external_ip and external_ip != "<none>":
                        # Detect whether to use HTTP or HTTPS
                        url_ports = ", ".join(
                            [
                                f"<{'https' if p['port'] == 443 else 'http'}://{external_ip}:{p['port']}|{external_ip}:{p['port']}/>"
                                for p in item["spec"].get("ports", [])
                            ]
                        )
                        url_ports_list.append(f"{name}: {url_ports}")

                    response_text += f"{name:<20}{svc_type:<15}{cluster_ip:<20}{external_ip:<20}{ports:<15}{age:<10}\n"
        else:
            response_text = f"Can't get {resource}.\n\n"

    response_text += "```\n"

    # Append URL ports list to response text
    if url_ports_list:
        response_text += "*Service URLs:*\n"
        for url_port in url_ports_list:
            response_text += f"- {url_port}\n"

    # Ensure response_text is JSON compatible
    response_text = (
        json.dumps(response_text)
        .encode()
        .decode("unicode_escape")
        .replace("\\n", "\n")
        .replace("\\", "")
        .replace('"', "")
    )

    logging.info("Response for namespace %s: %s", namespace, response_text)

    # Return the AI response to the user in Slack
    response = {
        "text": response_text,
        "attachments": [],
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
