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
)

const (
	serverInstructions = ``

	descriptionReadDocumentAsMarkdown = ``
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

	// Instantiate tool handlers.
	handler, err := tools.NewHandler()
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

	slog.Info("Starting server")
	// Run the server over stdin/stdout, until the client disconnects
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		slog.ErrorContext(ctx, "error in server.Run call", "error", err)
		panic("error in server.Run call: " + err.Error())
	}
}

func addTools(server *mcp.Server, handler *tools.Handler) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ReadDocumentAsMarkdown",
		Description: descriptionReadDocumentAsMarkdown,
	}, handler.ReadDocumentAsMarkdown)
}
