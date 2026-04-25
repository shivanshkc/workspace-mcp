package workspace

import (
	"context"
	"fmt"
	"io"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Client represents a client for Google Workspace services.
type Client struct {
	driveService *drive.Service
}

// NewClient creates a new client with the given config.
func NewClient(ctx context.Context, credentialsFilePath, tokenFilePath string, codeFunc CodeFunc) (*Client, error) {
	credentialsBytes, err := os.ReadFile(credentialsFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	authConfig, err := google.ConfigFromJSON(credentialsBytes, scopeDriveReadOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to parse client secret file to config: %w", err)
	}

	httpClient, err := createHttpClient(ctx, authConfig, tokenFilePath, codeFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	driveService, err := drive.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create drive service: %w", err)
	}

	return &Client{driveService: driveService}, nil
}

// DocumentSize holds the size metrics of a document.
type DocumentSize struct {
	CharCount int
}

// ReadDocumentAsMarkdown exports the Google Doc as markdown via the Drive API.
func (c *Client) ReadDocumentAsMarkdown(ctx context.Context, docID string) (string, error) {
	return c.exportMarkdown(ctx, docID)
}

// GetDocumentSize returns the character count of the Google Doc as exported to markdown.
func (c *Client) GetDocumentSize(ctx context.Context, docID string) (DocumentSize, error) {
	markdown, err := c.exportMarkdown(ctx, docID)
	if err != nil {
		return DocumentSize{}, err
	}
	return DocumentSize{CharCount: len(markdown)}, nil
}

func (c *Client) exportMarkdown(ctx context.Context, docID string) (string, error) {
	resp, err := c.driveService.Files.Export(docID, "text/markdown").Context(ctx).Download()
	if err != nil {
		return "", fmt.Errorf("failed to export document %s: %w", docID, err)
	}

	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read export response: %w", err)
	}

	return string(data), nil
}
