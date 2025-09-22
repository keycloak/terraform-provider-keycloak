package helper

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
)

var requiredEnvironmentVariables = []string{
	"KEYCLOAK_CLIENT_ID",
	"KEYCLOAK_URL",
	"KEYCLOAK_REALM",
}

var requiredEnvironmentVariablesForTokenAuth = []string{
	"KEYCLOAK_REALM",
	"KEYCLOAK_URL",
}

func CheckRequiredEnvironmentVariables(t *testing.T) {

	requiredEnvVars := requiredEnvironmentVariables
	if os.Getenv("KEYCLOAK_ACCESS_TOKEN") != "" {
		requiredEnvVars = requiredEnvironmentVariablesForTokenAuth
	}

	for _, requiredEnvironmentVariable := range requiredEnvVars {
		if value := os.Getenv(requiredEnvironmentVariable); value == "" {
			t.Fatalf("%s must be set before running acceptance tests.", requiredEnvironmentVariable)
		}
	}
}

func UpdateEnvFromTestEnvIfPresent() {

	testEnvFile := os.Getenv("TEST_ENV_FILE")
	if testEnvFile == "" {
		testEnvFile = "../test_env.json"
	}

	if _, err := os.Stat(testEnvFile); err == nil {
		fmt.Printf("Using %s to load environment variables...", testEnvFile)
		file, err := os.Open(testEnvFile)
		if err != nil {
			log.Fatalf("Unable to open env.json: %s", err)
		}
		defer file.Close()

		var envVars map[string]string
		if err := json.NewDecoder(file).Decode(&envVars); err != nil {
			log.Fatalf("Unable to decode env.json: %s", err)
		}

		for key, value := range envVars {
			if err := os.Setenv(key, value); err != nil {
				log.Fatalf("Unable to set environment variable %s: %s", key, err)
			}
		}
	}
}
