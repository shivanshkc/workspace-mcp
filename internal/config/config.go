package config

import (
	"encoding/json"
	"fmt"
	"os"
)

/*
Sample config JSON:
{
	"googleCredentialsFile": "/etc/workspace-mcp/credentials.shivanshbox.json",
	"googleTokenFile": "/etc/workspace-mcp/token.shivanshbox.json",
	"logFilePath": "/tmp/workspace-mcp.shivanshbox.log",
	"oauthCallbackPort": 47291
}
*/

// Config encapsulates all config required by the application.
type Config struct {
	GoogleCredentialsFile string `json:"googleCredentialsFile"`
	GoogleTokenFile       string `json:"googleTokenFile"`
	LogFilePath           string `json:"logFilePath"`
	OAuthCallbackPort     int    `json:"oauthCallbackPort"`
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
