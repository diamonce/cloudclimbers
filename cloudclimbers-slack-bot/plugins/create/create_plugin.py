from flask import Flask, request, jsonify

app = Flask(__name__)


@app.route("/create", methods=["POST"])
def create_environment():
    data = request.json
    # Implement the logic for creating the environment here
    response = {
        "text": "Environment created successfully!",
        "attachments": [{"text": "Details about the created environment..."}],
    }
    return jsonify(response)


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8081)
