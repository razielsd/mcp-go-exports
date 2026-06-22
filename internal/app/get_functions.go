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

var GetFunctionsTool = mcp.NewTool("get_functions",
	mcp.WithDescription("Get a list of functions in a package"),
	mcp.WithString("package",
		mcp.Description("The Go package to analyze"),
		mcp.Required(),
	),
	mcp.WithString("structName",
		mcp.Description("Optional: Filter by receiver structure name"),
	),
	mcp.WithString("functionName",
		mcp.Description("Optional: The specific function name to look for"),
	),
)

// HandleGetFunctions needs the loader to load and read the package data
func HandleGetFunctions(_ context.Context, loader *pkgmanager.PackageLoader, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pkg := mcp.ParseString(*request, "package", "")
	structName := mcp.ParseString(*request, "structName", "")
	fnName := mcp.ParseString(*request, "functionName", "")

	slog.Info("get_functions called", "package", pkg, "structName", structName, "functionName", fnName)

	if err := loader.Load(pkg); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to load package %s: %v", pkg, err)), nil
	}

	// Read the func.txt file
	path := filepath.Join(loader.SavePath, pkg, "func.txt")
	content, err := os.ReadFile(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("could not read functions for package %s: %v", pkg, err)), nil
	}

	resultText := string(content)
	lines := strings.Split(resultText, "\n")
	var filtered []string

	// Filter by structName if provided
	if structName != "" {
		for _, line := range lines {
			if strings.HasPrefix(line, structName+".") {
				filtered = append(filtered, line)
			}
		}
		resultText = strings.Join(filtered, "\n")
	}

	// Filter by functionName if provided
	if fnName != "" {
		var finalFiltered []string
		for _, line := range filtered { // If structName was provided, we filter the already filtered list
			if strings.Contains(line, fnName) {
				finalFiltered = append(finalFiltered, line)
			}
		}
		// Special case: if no structName was used, filtered is empty
		if structName == "" {
			for _, line := range lines {
				if strings.Contains(line, fnName) {
					finalFiltered = append(finalFiltered, line)
				}
			}
		} else if len(filtered) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No methods for struct %s found in package %s", structName, pkg)), nil
		}

		if len(finalFiltered) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("Function %s not found in package %s", fnName, pkg)), nil
		}
		resultText = strings.Join(finalFiltered, "\n")
	}

	return mcp.NewToolResultText(resultText), nil
}
