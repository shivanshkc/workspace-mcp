# Workspace MCP

An MCP server that exposes Google Workspace documents to AI agents. Currently supports reading Google Docs as markdown.

## Tools

### `ReadDocumentAsMarkdown`

Fetches a Google Doc and returns its content as markdown. Supports pagination via `limit` and `offset` for large documents.

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `doc_id`  | yes      | —       | Document ID from the Google Docs URL |
| `limit`   | no       | 1000    | Number of lines to return |
| `offset`  | no       | 0       | Line offset to start from |

## Setup

### Prerequisites

- Go 1.21+
- A [Google Cloud project](https://console.cloud.google.com) with the **Google Drive API** enabled
- OAuth 2.0 credentials (Desktop app type) downloaded as a JSON file

### Google Cloud configuration

1. Enable the **Google Drive API** for your project
2. Create OAuth 2.0 credentials (Desktop application) and download the credentials JSON
3. Add the `https://www.googleapis.com/auth/drive.readonly` scope to your OAuth consent screen

### Config file

Create a JSON config file (e.g. `/etc/workspace-mcp/config.json`):

```json
{
  "googleCredentialsFile": "/etc/workspace-mcp/credentials.json",
  "googleTokenFile": "/etc/workspace-mcp/token.json",
  "logFilePath": "/tmp/workspace-mcp.log"
}
```

### Build and install

```bash
make build
```

The binary is written to `bin/workspace-mcp`.

### Add to Claude Code

```bash
make claude
```

This builds the binary and registers it as an MCP server in Claude Code. On first run, the server will print an OAuth URL — open it in a browser, grant access, and paste the authorization code back into the terminal. The token is cached for subsequent runs.

To re-authorize (e.g. after changing scopes), delete the token file and restart.

## License

[MIT](LICENSE)
