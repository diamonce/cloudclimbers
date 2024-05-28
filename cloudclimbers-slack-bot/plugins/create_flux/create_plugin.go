package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gorilla/mux"
)

type RequestData struct {
	Commands  string            `json:"commands"`
	Variables map[string]string `json:"variables"`
	Hash      string            `json:"hash"`
	Payload   struct {
		State struct {
			Values map[string]map[string]struct {
				Value string `json:"value"`
			} `json:"values"`
		} `json:"state"`
	} `json:"payload"`
}

func executeCommand(command string) (string, string, error) {
	cmd := exec.Command("sh", "-c", command)
	stdout, err := cmd.Output()
	if err != nil {
		stderr := ""
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		return "", stderr, err
	}
	return string(stdout), "", nil
}

func createEnvironment(w http.ResponseWriter, r *http.Request) {
	var data RequestData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received request: %v", data)

	commands := data.Commands
	variables := data.Variables
	if variables == nil {
		variables = make(map[string]string)
	}

	userInputs := data.Payload.State.Values
	for blockID, blockValue := range userInputs {
		for _, value := range blockValue {
			variables[blockID] = value.Value
		}
	}

	missingVariables := make(map[string]string)
	for key, value := range variables {
		if value == "" {
			missingVariables[key] = ""
		}
	}

	if len(missingVariables) > 0 {
		inputBlocks := []map[string]interface{}{}
		for varName := range missingVariables {
			inputBlocks = append(inputBlocks, map[string]interface{}{
				"type":     "input",
				"block_id": varName,
				"element":  map[string]interface{}{"type": "plain_text_input", "action_id": varName, "placeholder": map[string]interface{}{"type": "plain_text", "text": "Enter " + varName}},
				"label":    map[string]interface{}{"type": "plain_text", "text": varName},
			})
		}

		response := map[string]interface{}{
			"text":   "Please provide the following variables:",
			"blocks": inputBlocks,
			"buttons": []map[string]interface{}{
				{"type": "button", "text": "Submit Variables", "action_id": "create_environment_flux"},
			},
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	for key, value := range variables {
		commands = strings.ReplaceAll(commands, "${"+strings.ToUpper(key)+"}", value)
	}

	log.Printf("Commands to be executed: %s", commands)

	stdout, stderr, err := executeCommand(commands)
	if err != nil {
		log.Printf("Command execution failed: %s", stderr)
		http.Error(w, stderr, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"text": "Environment created successfully!",
		"attachments": []map[string]interface{}{
			{"text": "Details about the created environment: " + stdout},
		},
		"buttons": []map[string]interface{}{
			{"type": "button", "text": "Get Environment Status", "action_id": "get_environment_status"},
		},
	}
	json.NewEncoder(w).Encode(response)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/create", createEnvironment).Methods("POST")

	log.Fatal(http.ListenAndServe(":8085", r))
}
