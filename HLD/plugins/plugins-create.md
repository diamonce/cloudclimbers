very first draft

## Create plugins
Any programming language is suitable for creating a plugin.
The main condition is that the plugin must use a web server to listen to the port specified in the settings, to which the BOT will send information from the user.

1) Place the plugin in a folder with other plugins.
```code
from flask import Flask, request, jsonify

app = Flask(__name__)


@app.route("/get", methods=["POST"])
def get_environment():
    data = request.json
    # Implement the logic for getting the environment status here

    response = {
        "text": "Environment status retrieved successfully!",
        "attachments": [{"text": "Details about the environment status..."}],
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
```

2) Create a section in the configuration file using a template.

```code
slack_app_token: ""
slack_bot_token: ""
mongo_uri: "mongodb://mongodb:27017"  # Ensure this is correctly formatted
database_name: "cloudclimbers_db"

plugins:

  create_flux:
    url: "http://create-flux-plugin:8081/create"
    buttons:
      - text: "Create Environment with Flux"
        action_id: "create_environment_flux"
        emoji: ":sparkles:"  # Emoji
  get:
    url: "http://get-plugin:8082/get"
    buttons:
      - text: "Get Environment Status"
        action_id: "get_environment_status"
        emoji: ":mag:"  # Emoji for getting status
    commands: |
      kubectl get pods -n ${NAMESPACE} -o wide
      kubectl describe pods -n ${NAMESPACE}
    variables:
      NAMESPACE: "value1"
      DEPLOYMENT_NAME: "value2"
    hash:
      type: "SHA-256"
      value: "" # Will be calculated


main_buttons:
  - text: "Main Menu"
    action_id: "list_enabled_plugins"
    emoji: ":rocket:"  # Existing emoji for listing plugins
  - text: "Help"
    action_id: "help"
    emoji: ":information_source:"  # Existing emoji for help
```

3) Write Dockerfile
```code
FROM python:3.9-alpine

WORKDIR /app

COPY get_plugin.py .

RUN pip install Flask

EXPOSE 8082

CMD ["python", "get_plugin.py"]
```

4) Encrypt the resulting configuration file.
```code
openssl enc -aes-256-cbc -salt -in config.yaml -out config.yaml.enc -k "<your_password>"
kubectl delete secret slack-bot-config  -n cloudclimbers
kubectl create secret generic slack-bot-config --from-literal=encryptionPassword=<your_password> --from-file=config.yaml.enc=config.yaml.enc -n cloudclimbers
```
6) Add information on assembling the plugin to the makefile.

```code
# Build Docker images
docker-build: build
	echo "==> Building Docker images..."
	docker buildx build --platform $(OS)/$(ARCH) -t $(MAIN_IMAGE_REPO):$(IMAGE_TAG) .
	docker buildx build --platform $(OS)/$(ARCH) -t $(FLUX_CREATE_IMAGE_REPO):$(IMAGE_TAG) $(FLUX_CREATE_PLUGIN_DIR)
	docker buildx build --platform $(OS)/$(ARCH) -t $(GET_IMAGE_REPO):$(IMAGE_TAG) $(GET_PLUGIN_DIR)
	
	echo "==> Docker build completed"

# Push Docker images to GCR
docker-push: gcr-init docker-build
	echo "==> Pushing Docker images to GCR..."
	docker push $(MAIN_IMAGE_REPO):$(IMAGE_TAG)
	docker push $(FLUX_CREATE_IMAGE_REPO):$(IMAGE_TAG)
	docker push $(GET_IMAGE_REPO):$(IMAGE_TAG)

	echo "==> Docker images pushed to GCR"
```
6) Build the project and place everything on the selected server.
