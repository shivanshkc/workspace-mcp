package config

import (
	"encoding/json"
	"fmt"
	"os"
)

/*
Sample config JSON:
{
	"logFilePath": "/tmp/workspace-mcp.shivanshbox.log"
}
*/

// Config encapsulates all config required by the application.
type Config struct {
	LogFilePath string `json:"logFilePath"`
}

// Load configs from the given JSON file.
func Load(jsonPath string) (Config, error) {
	content, err := os.ReadFile(jsonPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file at %s: %w", jsonPath, err)
	}

	var configs Config
	if err := json.Unmarshal(content, &configs); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config file at %s: %w", jsonPath, err)
	}

	return configs, nil
}
