package helper

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

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
