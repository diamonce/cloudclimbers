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
