import shlex
from flask import Flask, request, jsonify
import logging
import subprocess

app = Flask(__name__)

# Configure logging
logging.basicConfig(level=logging.INFO)

def execute_commands(commands):
    all_stdout = []
    all_stderr = []

    for command in commands:
        try:
            # Разбиваем команду на список аргументов, чтобы не использовать "Shell=True"
            args = shlex.split(command)
            # Выполняем команду с безопасной передачей аргументов
            result = subprocess.run(
                args,
                check=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
            )
            stdout = result.stdout.decode("utf-8")
            stderr = result.stderr.decode("utf-8")

            if stdout:
                print("Output:")
                print(stdout)
            if stderr:
                print("Errors:")
                print(stderr)

            all_stdout.append(stdout)
            all_stderr.append(stderr)
        except subprocess.CalledProcessError as e:
            print(f"Command '{command}' failed with return code {e.returncode}")
            print(f"Error output: {e.stderr.decode('utf-8')}")
            all_stderr.append(e.stderr.decode("utf-8"))
            break
        except Exception as e:
            print(f"An unexpected error occurred: {str(e)}")
            all_stderr.append(str(e))
            break

    return "\n".join(all_stdout), "\n".join(all_stderr)

@app.route("/create", methods=["POST"])
def create_environment():
    data = request.json

    logging.info("Received request: %s", data)

    commands = data.get("commands", "")
    variables = data.get("variables", {})
    hash_value = data.get("hash", {})

    logging.info("Commands: %s", commands)
    logging.info("Variables: %s", variables)
    logging.info("Hash: %s", hash_value)

    if variables is None:
        variables = {}

    payload = data.get("payload", {})
    user_inputs = payload.get("state", {}).get("values", {})

    logging.info("User inputs: %s", user_inputs)

    for block_id, block_value in user_inputs.items():
        action_id = list(block_value.keys())[0]
        variables[block_id] = block_value[action_id].get("value", "")

    logging.info("Updated Variables: %s", variables)

    missing_variables = {key: "" for key, value in variables.items() if value == ""}

    if missing_variables:
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
                    "action_id": "create_environment",
                }
            ],
        }
        return jsonify(response)

    for key, value in variables.items():
        commands = commands.replace(f"${{{key.upper()}}}", value)

    logging.info("Commands to be executed: %s", commands)

    # Преобразуем команды в список строк
    command_list = commands.split("\n")
    stdout, stderr = execute_commands(command_list)

    response = {
        "text": "Environment created successfully!",
        "attachments": [
            {"text": "Details about the created environment...\n" + stdout}
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
