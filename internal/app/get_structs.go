package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/razielsd/mcp-go-exports/internal/pkgmanager"
)

var GetStructsTool = mcp.NewTool("get_structs",
	mcp.WithDescription("Get a list of structs in a package"),
	mcp.WithString("package",
		mcp.Description("The Go package to analyze"),
		mcp.Required(),
	),
	mcp.WithString("structName",
		mcp.Description("Optional: The specific struct name to look for"),
	),
)

func HandleGetStructs(_ context.Context, loader *pkgmanager.PackageLoader, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pkg := mcp.ParseString(*request, "package", "")
	structName := mcp.ParseString(*request, "structName", "")

	slog.Info("get_structs called", "package", pkg, "structName", structName)

	if err := loader.Load(pkg); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to load package %s: %v", pkg, err)), nil
	}

	path := filepath.Join(loader.SavePath, pkg, "struct.txt")
	content, err := os.ReadFile(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("could not read structs for package %s: %v", pkg, err)), nil
	}

	resultText := string(content)
	if structName != "" {
		lines := strings.Split(resultText, "\n")
		var filtered []string
		for i := 0; i < len(lines); i++ {
			line := lines[i]
			if strings.HasPrefix(line, structName+" {") {
				start := i
				for i+1 < len(lines) && strings.HasPrefix(lines[i+1], "  ") {
					i++
				}
				filtered = append(filtered, strings.Join(lines[start:i+1], "\n"))
			}
		}
		if len(filtered) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("Struct %s not found in package %s", structName, pkg)), nil
		}
		return mcp.NewToolResultText(strings.Join(filtered, "\n\n")), nil
	}

	return mcp.NewToolResultText(resultText), nil
}
