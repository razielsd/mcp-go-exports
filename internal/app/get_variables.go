package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/razielsd/mcp-go-exports/internal/pkgmanager"
)

var GetVariablesTool = mcp.NewTool("get_variables",
	mcp.WithDescription("Get a list of variables in a package"),
	mcp.WithString("package",
		mcp.Description("The Go package to analyze"),
		mcp.Required(),
	),
)

func HandleGetVariables(_ context.Context, loader *pkgmanager.PackageLoader, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pkg := mcp.ParseString(*request, "package", "")

	slog.Info("get_variables called", "package", pkg)

	if err := loader.Load(pkg); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to load package %s: %v", pkg, err)), nil
	}

	path := filepath.Join(loader.SavePath, pkg, "var.txt")
	content, err := os.ReadFile(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("could not read variables for package %s: %v", pkg, err)), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}
