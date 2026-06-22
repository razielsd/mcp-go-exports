package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/razielsd/mcp-go-exports/internal/app"
	"github.com/razielsd/mcp-go-exports/internal/config"
	"github.com/razielsd/mcp-go-exports/internal/pkgmanager"
)

func main() {
	// Initialize slog logger to write to stderr (so it doesn't interfere with MCP stdio)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	// Load configuration for the data path
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}
	savePath := cfg.GetDataDir()

	// Initialize the package loader with the configured paths
	loader := pkgmanager.NewPackageLoader(savePath, cfg.Local)

	// Create a new MCP server
	s := server.NewMCPServer(
		"mcp-go-exports",
		"1.0.0",
		server.WithLogger(logger),
		server.WithRecovery(),
	)

	// Register tools using closures to inject the loader dependency
	s.AddTool(app.EchoTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return app.HandleEcho(ctx, &request)
	})

	// Use closures to inject the loader into handlers that need it
	s.AddTool(app.GetFunctionsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return app.HandleGetFunctions(ctx, loader, &request)
	})

	s.AddTool(app.GetConstantsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return app.HandleGetConstants(ctx, loader, &request)
	})

	s.AddTool(app.GetStructsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return app.HandleGetStructs(ctx, loader, &request)
	})

	s.AddTool(app.GetVariablesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return app.HandleGetVariables(ctx, loader, &request)
	})

	switch cfg.Default.Transport {
	case "sse":
		sseServer := server.NewSSEServer(s)
		slog.Info("SSE server starting", slog.String("host", cfg.Transport.Sse))
		if err := sseServer.Start(cfg.Transport.Sse); err != nil {
			slog.Error("SSE server error", "error", err)
		}
	case "http":
		httpServer := server.NewStreamableHTTPServer(s)
		slog.Info("http streamable server starting", slog.String("host", cfg.Transport.HTTP))
		if err := httpServer.Start(cfg.Transport.HTTP); err != nil {
			slog.Error("StreamableHTTP server error", "error", err)
		}
	default:
		if err := server.ServeStdio(s); err != nil {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}
}
