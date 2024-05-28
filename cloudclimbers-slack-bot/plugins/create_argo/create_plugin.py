from flask import Flask, request, jsonify
import logging
import subprocess

app = Flask(__name__)

# Configure logging
logging.basicConfig(level=logging.INFO)


def execute_command(command):
    print("function to execute commands is here")
    try:
        # Run the command and capture the output and errors
        result = subprocess.run(
            command,
            check=True,
            shell=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
        # Decode the output and errors from bytes to string
        stdout = result.stdout.decode("utf-8")
        stderr = result.stderr.decode("utf-8")

        # Print the standard output
        if stdout:
            print("Output:")
            print(stdout)

        # Print the standard error
        if stderr:
            print("Errors:")
            print(stderr)

        return stdout, stderr
    except subprocess.CalledProcessError as e:
        # Handle the error
        print(f"Command '{command}' failed with return code {e.returncode}")
        print(f"Error output: {e.stderr.decode('utf-8')}")
        return None, e.stderr.decode("utf-8")
    except Exception as e:
        # Handle unexpected exceptions
        print(f"An unexpected error occurred: {str(e)}")
        return None, str(e)


@app.route("/create", methods=["POST"])
def create_environment():
    data = request.json

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
                    "action_id": "create_environment_argo",
                }
            ],
        }
        return jsonify(response)

    # Replace placeholders in commands with user inputs, considering uppercase placeholders
    for key, value in variables.items():
        commands = commands.replace(f"${{{key.upper()}}}", value)

    # Log the final commands to be executed
    logging.info("Commands to be executed: %s", commands)

    # Implement the logic for creating the environment here
    # For demonstration, just printing the commands
    print("Commands to be executed:", commands)

    # os.exec(commands)
    stdout, stderr = execute_command(commands)

    response = {
        "text": "Environment created successfully!",
        "attachments": [
            {"text": "Details about the created environment..." + str(stdout)}
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
    app.run(host="0.0.0.0", port=8081)
