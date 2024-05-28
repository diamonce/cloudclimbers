import requests
from flask import Flask, request, jsonify
app = Flask(__name__)

# Connection params
token_path = "/path/to/token"
namespace = "monitoring"
api_server = "https://your-kubernetes-api-server"
# for production need to set valid cert
#ca_cert_path = "/path/to/ca.crt"
ca_cert_path = False


# Read token from file
def read_token(token_path):
    with open(token_path, 'r') as token_file:
        return token_file.read().strip()
    
# Function for make requests
def get_resources(url, headers, ca_cert_path):
    response = requests.get(url, headers=headers, verify=ca_cert_path)
    if response.status_code == 200:
        return response.json()
    else:
        print(f"Error getting resources: {response.status_code} - {response.text}")
        return None

@app.route("/get", methods=["POST"])
def get_environment():
    data = request.json
    # Implement the logic for getting the environment status here

    # Read token
    token = read_token(token_path)

    # Build headers
    headers = {
        "Authorization": f"Bearer {token}",
        "Accept": "application/json"
    }

    # URLS to get all pods, replicasets, services and other resources in namespace
    urls = {
        "pods": f"{api_server}/api/v1/namespaces/{namespace}/pods",
        "replicasets": f"{api_server}/apis/apps/v1/namespaces/{namespace}/replicasets",
        "deployments": f"{api_server}/apis/apps/v1/namespaces/{namespace}/deployments",
        "services": f"{api_server}/api/v1/namespaces/{namespace}/services"
    }


    # Getting all resources
    resources = {resource: get_resources(url, headers, ca_cert_path) for resource, url in urls.items()}

    # Making answer
    response_text = ""
    for resource, data in resources.items():
        if data:
            resource_names = [item['metadata']['name'] for item in data['items']]
            response_text += f"{resource.capitalize()}:\n" + "\n".join(resource_names) + "\n\n"
        else:
            response_text += f"Can't get {resource}.\n\n"

    response = {
        "text": "Environment status retrieved successfully!",
        "attachments": [{"text": f"Resources in namespace {namespace}:\n{response_text}"}],
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
