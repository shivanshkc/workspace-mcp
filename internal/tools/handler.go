package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/docs/v1"
)

// ReadDocumentAsMarkdownInput holds the input parameters required by the ReadDocumentAsMarkdown tool.
type ReadDocumentAsMarkdownInput struct {
	DocID  string
	Limit  int
	Offset int
}

type ReadDocumentAsMarkdownOutput struct {
	Content string
}

// Handler encapsulates all tool handler methods.
type Handler struct {
	docService *docs.Service
}

// NewHandler instantiates a new Handler instance.
func NewHandler() (*Handler, error) {
	return &Handler{}, nil
}

// ReadDocumentAsMarkdown reads the specified Google Doc from start to end, converts it to markdown,
// and returns the required number of lines as specified by limit and offset.
//
// The default values for limit and offset are 1000 and zero respectively.
//
// The returned text is formatted as cat -n output, meaning each line prefixed with its line number
// and a tab, starting at line 1. For example:
//
//	1        first line of file
//	2        second line of file
//
// This tool's signature is meant to be similar to Claude Code's built-in Read tool.
func (h *Handler) ReadDocumentAsMarkdown(
	ctx context.Context, req *mcp.CallToolRequest, input ReadDocumentAsMarkdownInput,
) (*mcp.CallToolResult, ReadDocumentAsMarkdownOutput, error) {
	return nil, ReadDocumentAsMarkdownOutput{}, nil
}
