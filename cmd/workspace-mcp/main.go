package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/shivanshkc/workspacemcp/internal/config"
	"github.com/shivanshkc/workspacemcp/internal/tools"
	"github.com/shivanshkc/workspacemcp/pkg/workspace"
)

const (
	serverInstructions = `You are connected to a Google Workspace MCP server. Use it to read content from Google Docs on behalf of the user.

Rules:
- Always use the document ID from the URL, not the document title.
- Before reading an unfamiliar document, call GetDocumentSize to check its character count.
- Prefer reading documents in chunks using offset and limit rather than all at once.
- Do not infer or fabricate document content — only use what the tool returns.`

	descriptionGetDocumentSize = `Returns the size of a Google Doc in characters (as exported to markdown).

Use this before reading an unfamiliar document to decide how to page through it.

Parameters:
- doc_id (required): the document ID from the Google Docs URL`

	descriptionReadDocumentAsMarkdown = `Fetches a Google Doc and returns its content as markdown.

offset and limit are in characters. To page through a document, advance offset by limit on each call.
Omitting limit returns the entire document. If you're unsure about the size of the document, use GetDocumentSize first.

Parameters:
- doc_id (required): the document ID from the Google Docs URL (e.g. "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms")
- limit (optional): number of characters to return; omit to read the entire document
- offset (optional): character offset to start from, default 0`
)

func main() {
	// Root context of the server.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// This allows the server to be easily used with different environments. Example:
	// $ app -config config.stage.json
	configPath := flag.String("config", "/etc/workspace-mcp/config.json", "path to config file")
	flag.Parse()

	conf, err := config.Load(*configPath)
	if err != nil {
		// Log will not be visible in the log file.
		panic("failed to load configs: " + err.Error())
	}

	// Server will log to this file.
	file, err := os.OpenFile(conf.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Log will not be visible in the log file.
		panic("failed to open log file")
	}

	defer func() { _ = file.Close() }()

	// Use slog to log to the file.
	slog.SetDefault(slog.New(slog.NewTextHandler(file, nil)))
	slog.InfoContext(ctx, fmt.Sprintf("---------- NEW RUN: %s ----------", time.Now().Format(time.RFC822)))

	// Establish connectivity with Google Workspace.
	wClient, err := workspace.NewClient(ctx, conf.GoogleCredentialsFile, conf.GoogleTokenFile, codeFunc)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create workspace client", "error", err)
		panic("failed to create workspace client: " + err.Error())
	}

	slog.InfoContext(ctx, "connected to google workspace successfully")

	// Instantiate tool handlers.
	handler, err := tools.NewHandler(wClient)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create handler", "error", err)
		panic("failed to create handler: " + err.Error())
	}

	// Create server with instructions. Tools will be attached separately.
	server := mcp.NewServer(
		&mcp.Implementation{Name: "Workspace MCP", Version: "v0.0.0"},
		&mcp.ServerOptions{Instructions: serverInstructions})

	// Attach all tools.
	addTools(server, handler)

	slog.InfoContext(ctx, "starting server...")
	// Run the server over stdin/stdout, until the client disconnects
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		slog.ErrorContext(ctx, "error in server.Run call", "error", err)
		panic("error in server.Run call: " + err.Error())
	}
}

func addTools(server *mcp.Server, handler *tools.Handler) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "GetDocumentSize",
		Description: descriptionGetDocumentSize,
	}, handler.GetDocumentSize)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ReadDocumentAsMarkdown",
		Description: descriptionReadDocumentAsMarkdown,
	}, handler.ReadDocumentAsMarkdown)
}

func codeFunc(authURL string) (string, error) {
	fmt.Println("Go to the following link in your browser then type the authorization code:\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return "", fmt.Errorf("failed to scan authorization code: %w", err)
	}

	return authCode, nil
}
