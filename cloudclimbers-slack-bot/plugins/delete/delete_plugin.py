from flask import Flask, request, jsonify

app = Flask(__name__)


@app.route("/delete", methods=["POST"])
def delete_environment():
    data = request.json
    # Implement the logic for deleting the environment here

    response = {
        "text": "Environment deleted successfully!",
        "attachments": [{"text": "Details about the deleted environment..."}],
    }
    return jsonify(response)


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8083)
