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

var GetConstantsTool = mcp.NewTool("get_constants",
	mcp.WithDescription("Get a list of constants in a package"),
	mcp.WithString("package",
		mcp.Description("The Go package to analyze"),
		mcp.Required(),
	),
)

func HandleGetConstants(_ context.Context, loader *pkgmanager.PackageLoader, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pkg := mcp.ParseString(*request, "package", "")

	slog.Info("get_constants called", "package", pkg)

	if err := loader.Load(pkg); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to load package %s: %v", pkg, err)), nil
	}

	path := filepath.Join(loader.SavePath, pkg, "const.txt")
	content, err := os.ReadFile(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("could not read constants for package %s: %v", pkg, err)), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}
