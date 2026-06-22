package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
)

// EchoTool defines the echo tool metadata
var EchoTool = mcp.NewTool("echo",
	mcp.WithDescription("Echoes back the input text"),
	mcp.WithString("message",
		mcp.Description("The message to echo"),
		mcp.Required(),
	),
)

// HandleEcho processes the echo tool logic
func HandleEcho(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	slog.Info("Handling echo tool call", "request", request)

	msg := mcp.ParseString(*request, "message", "")
	if msg == "" {
		slog.Warn("Echo tool called with missing message", "request", request)
		return mcp.NewToolResultError("missing or invalid argument 'message'"), nil
	}

	slog.Debug("Successfully processed echo tool call", "message", msg)
	return mcp.NewToolResultText(fmt.Sprintf("Echo: %s", msg)), nil
}
