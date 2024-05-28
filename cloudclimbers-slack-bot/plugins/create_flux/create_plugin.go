package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	//	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	// "sigs.k8s.io/controller-runtime/pkg/scheme"
	// ctrl "sigs.k8s.io/controller-runtime"
)

type RequestData struct {
	Commands  string            `json:"commands"`
	Variables map[string]string `json:"variables"`
	Hash      json.RawMessage   `json:"hash"`
	Payload   struct {
		State struct {
			Values map[string]map[string]struct {
				Value string `json:"value"`
			} `json:"values"`
		} `json:"state"`
	} `json:"payload"`
}

func executeCommand(logger *zap.Logger, command string) (string, string, error) {
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

func createEnvironment(logger *zap.Logger, w http.ResponseWriter, r *http.Request) {
	logger.Info("createEnvironment function started")

	var data RequestData

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		logger.Info(fmt.Sprintf("%v", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Info("Received request", zap.Any("data", data))

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
				"element": map[string]interface{}{
					"type":        "plain_text_input",
					"action_id":   varName,
					"placeholder": map[string]interface{}{"type": "plain_text", "text": "Enter " + varName},
				},
				"label": map[string]interface{}{"type": "plain_text", "text": varName},
			})
		}

		response := map[string]interface{}{
			"text":   "Please provide the following variables:",
			"blocks": inputBlocks,
			"buttons": []map[string]interface{}{
				{"type": "button", "text": "Submit Variables", "action_id": "create_environment_flux"},
			},
		}
		logger.Info("Missing variables, prompting user for input", zap.Any("response", response))
		json.NewEncoder(w).Encode(response)
		return
	}

	for key, value := range variables {
		commands = strings.ReplaceAll(commands, "${"+strings.ToUpper(key)+"}", value)
	}

	logger.Info("Commands to be executed", zap.String("commands", commands))

	stdout, stderr, err := executeCommand(logger, commands)
	if err != nil {
		logger.Error("Command execution failed", zap.String("stderr", stderr))
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
	logger.Info("Environment created successfully", zap.Any("response", response))
	json.NewEncoder(w).Encode(response)
}

func listFluxApps(logger *zap.Logger, w http.ResponseWriter, r *http.Request) {
	logger.Info("listFluxApps function started")

	cfg, err := rest.InClusterConfig()
	if err != nil {
		logger.Error("Failed to get in-cluster config", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	scheme := runtime.NewScheme()
	helmv2.AddToScheme(scheme)

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		logger.Error("Failed to create Kubernetes client", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var helmReleases helmv2.HelmReleaseList
	if err := k8sClient.List(context.Background(), &helmReleases); err != nil {
		logger.Error("Failed to list HelmReleases", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	releaseNames := []string{}
	for _, hr := range helmReleases.Items {
		releaseNames = append(releaseNames, hr.Name)
	}

	response := map[string]interface{}{
		"text": "Available Applications: " + strings.Join(releaseNames, ", "),
		"attachments": []map[string]interface{}{
			{"text": "Helm Releases: " + strings.Join(releaseNames, ", ")},
		},
		"buttons": []map[string]interface{}{
			{"type": "button", "text": "Deploy Environment", "action_id": "deploy_environment"},
		},
	}

	logger.Info("HelmReleases listed successfully", zap.Any("response", response))
	json.NewEncoder(w).Encode(response)
}

func loggingMiddleware(logger *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("Request received", zap.String("method", r.Method), zap.String("url", r.URL.String()))
			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic("Failed to initialize zap logger")
	}
	defer logger.Sync() // flushes buffer, if any

	r := mux.NewRouter()
	r.Use(loggingMiddleware(logger))
	r.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		listFluxApps(logger, w, r)
		//createEnvironment(logger, w, r)
	}).Methods("POST")
	r.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		listFluxApps(logger, w, r)
	}).Methods("GET")

	logger.Info("Starting server on :8085")
	if err := http.ListenAndServe(":8085", r); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
