package workspace

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

// Client represents a client for the Docs, Slides and Sheets services.
type Client struct {
	docService *docs.Service
}

// NewClient creates a new client with the given config.
func NewClient(ctx context.Context, credentialsFilePath, tokenFilePath string, codeFunc CodeFunc) (*Client, error) {
	credentialsBytes, err := os.ReadFile(credentialsFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	authConfig, err := google.ConfigFromJSON(credentialsBytes, scopeDocsReadyOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to parse client secret file to config: %w", err)
	}

	httpClient, err := createHttpClient(ctx, authConfig, tokenFilePath, codeFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	docService, err := docs.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create docs client: %w", err)
	}

	return &Client{docService: docService}, nil
}
