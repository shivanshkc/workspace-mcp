package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

const (
	scopeDocsReadyOnly = "https://www.googleapis.com/auth/documents.readonly"
)

// CodeFunc accepts the auth URL and returns the OAuth code or an error.
type CodeFunc func(authURL string) (string, error)

// createHttpClient creates an http client using the given oauth config.
func createHttpClient(ctx context.Context, config *oauth2.Config, tokenFilePath string, codeFunc CodeFunc) (*http.Client, error) {
	slog.InfoContext(ctx, "fetching token from local file")

	// If token found in local file, use it.
	token, err := tokenFromFile(tokenFilePath)
	if err == nil {
		return config.Client(context.Background(), token), nil
	}

	slog.ErrorContext(ctx, "failed to get token from local file", "error", err)
	slog.InfoContext(ctx, "fetching token from the web")

	// Token not found in local file, get it from web.
	token, err = tokenFromWeb(ctx, config, codeFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to get token from the web: %w", err)
	}

	// Save token so network call can be avoided next time.
	if err := saveToken(tokenFilePath, token); err != nil {
		slog.ErrorContext(ctx, "failed to save token to file", "error", err)
	}

	return config.Client(ctx, token), nil
}

// tokenFromFile retrieves token from a local file, if present.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() { _ = f.Close() }()

	var token oauth2.Token
	if err := json.NewDecoder(f).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to read token from file: %w", err)
	}

	return &token, nil
}

// tokenFromWeb requests a token from the web.
func tokenFromWeb(ctx context.Context, config *oauth2.Config, codeFunc CodeFunc) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	// Obtain auth code from the URL by calling the package user's logic.
	authCode, err := codeFunc(authURL)
	if err != nil {
		return nil, fmt.Errorf("codeFunc returned error: %w", err)
	}

	token, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve token from web: %w", err)
	}

	return token, nil
}

// saveToken saves the token to the given local file.
func saveToken(tokenFilePath string, token *oauth2.Token) error {
	f, err := os.OpenFile(tokenFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file for writing token: %w", err)
	}

	defer func() { _ = f.Close() }()

	if err := json.NewEncoder(f).Encode(token); err != nil {
		return fmt.Errorf("failed to write token to file: %w", err)
	}

	return nil
}
