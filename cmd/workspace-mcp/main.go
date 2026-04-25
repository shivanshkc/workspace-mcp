package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/shivanshkc/workspacemcp/internal/config"
	"github.com/shivanshkc/workspacemcp/internal/tools"
	"github.com/shivanshkc/workspacemcp/pkg/workspace"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverInstructions = `You are connected to a Google Workspace MCP server. Use it to read content from Google Docs and Google Sheets on behalf of the user.

Rules (Docs):
- Always use the document ID from the URL, not the document title.
- Before reading an unfamiliar document, call GetDocumentSize to check its character count.
- Prefer reading documents in chunks using offset and limit rather than all at once.
- Do not infer or fabricate document content — only use what the tool returns.

Rules (Sheets):
- Always call ListSheets first to discover sheet names before reading any data.
- Call GetSheetSize before ReadSheet to understand the data extent and plan your ranges.
- Read in focused ranges rather than fetching the whole sheet at once.
- Sheet names are case-sensitive and must exactly match what ListSheets returns.
- A sheet may contain multiple independent tables at different positions — read in small ranges first to locate structure before reading large blocks.
- Do not infer or fabricate sheet data — only use what the tool returns.`

	descriptionListSheets = `Returns the names of all sheets (tabs) in a Google Spreadsheet.

Always call this first when working with an unfamiliar spreadsheet.
The names returned are the exact strings required by GetSheetSize and ReadSheet.

Parameters:
- spreadsheet_id (required): the spreadsheet ID from the URL`

	descriptionGetSheetSize = `Returns the extent of actual data in a sheet — the last populated row and column.

Use this before ReadSheet to understand how much data exists and plan your ranges.
dataRange in the response is ready to pass directly to ReadSheet as the range parameter.

Note: this fetches the entire sheet internally to compute the extent.
Avoid calling it repeatedly on very large sheets.

Parameters:
- spreadsheet_id (required): the spreadsheet ID from the URL
- sheet_name (required): exact name as returned by ListSheets`

	descriptionReadSheet = `Reads cells from a Google Sheet and returns them as a 2D array (array of rows, each row an array of values).

Response format:
- Values are unformatted: numbers are numbers, booleans are booleans, strings are strings.
- Formula cells return their computed value, not the formula itself.
- Empty cells mid-row are returned as empty string "".
- Trailing empty cells in a row are omitted, so rows may have different lengths.
- A shorter row means no data beyond its last value — not an error.
- If range is omitted, the full data extent is fetched automatically via GetSheetSize.
- Only omit range after confirming via GetSheetSize that the sheet fits in your context window.

A1 notation guide:
- "A1:E10" — columns A to E, rows 1 to 10
- "B2:D50" — start mid-sheet at row 2, column B
- "A1:A100" — single column
- Columns beyond Z use two letters: AA, AB, ..., AZ, BA, ...

Parameters:
- spreadsheet_id (required): the spreadsheet ID from the URL
- sheet_name (required): exact name as returned by ListSheets
- range (optional): A1 notation; if omitted reads the full data extent via GetSheetSize`

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
	wClient, err := workspace.NewClient(ctx, conf.GoogleCredentialsFile, conf.GoogleTokenFile, conf.OAuthCallbackPort, autoCodeFunc(ctx, conf.OAuthCallbackPort))
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
	// Docs tools.
	mcp.AddTool(server, &mcp.Tool{
		Name:        "GetDocumentSize",
		Description: descriptionGetDocumentSize,
	}, handler.GetDocumentSize)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ReadDocumentAsMarkdown",
		Description: descriptionReadDocumentAsMarkdown,
	}, handler.ReadDocumentAsMarkdown)

	// Sheets tools.
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ListSheets",
		Description: descriptionListSheets,
	}, handler.ListSheets)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "GetSheetSize",
		Description: descriptionGetSheetSize,
	}, handler.GetSheetSize)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ReadSheet",
		Description: descriptionReadSheet,
	}, handler.ReadSheet)
}

// autoCodeFunc returns a CodeFunc that starts a local HTTP server on the given port,
// opens the browser to the auth URL, and captures the code when Google redirects back.
func autoCodeFunc(ctx context.Context, port int) workspace.CodeFunc {
	return func(authURL string) (string, error) {
		codeCh := make(chan string, 1)

		srv := &http.Server{
			Addr: fmt.Sprintf(":%d", port),
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				code := r.URL.Query().Get("code")
				if code == "" {
					http.Error(w, "missing code", http.StatusBadRequest)
					return
				}
				_, _ = fmt.Fprintln(w, "Authorization complete. You may close this tab.")
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				codeCh <- code
			}),
		}

		go func() { _ = srv.ListenAndServe() }()

		if err := exec.Command("open", authURL).Start(); err != nil {
			_ = srv.Close()
			return "", fmt.Errorf("failed to open browser: %w", err)
		}

		select {
		case code := <-codeCh:
			_ = srv.Close()
			return code, nil
		case <-ctx.Done():
			_ = srv.Close()
			return "", ctx.Err()
		}
	}
}

// manualCodeFunc is a CodeFunc that prints the auth URL and reads the code from stdin.
//
//nolint:unused
func manualCodeFunc(authURL string) (string, error) {
	fmt.Println("Go to the following link in your browser then type the authorization code:\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return "", fmt.Errorf("failed to scan authorization code: %w", err)
	}

	return authCode, nil
}
