package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/razielsd/mcp-go-exports/internal/config"
	"github.com/razielsd/mcp-go-exports/internal/pkgmanager"
)

func main() {
	const minArgs = 2
	if len(os.Args) < minArgs {
		fmt.Println("Usage: go run cmd/cli/main.go <package_name>")
		fmt.Println("Example: go run cmd/cli/main.go fmt")
		os.Exit(1)
	}

	packageName := os.Args[1]
	// Sanitize input to prevent log injection
	sanitizedPackageName := strings.ReplaceAll(packageName, "\n", "")
	sanitizedPackageName = strings.ReplaceAll(sanitizedPackageName, "\r", "")

	// Load configuration from config.json
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}
	savePath := cfg.GetDataDir()
	localCache := cfg.Local

	fmt.Printf("Initializing PackageLoader with cache at: %s\n", savePath)
	fmt.Printf("Using local Go module cache: %s\n", localCache)
	loader := pkgmanager.NewPackageLoader(savePath, localCache)

	fmt.Printf("Processing package: %s...\n", packageName)
	err = loader.Load(packageName)
	if err != nil {
		slog.Error("Error loading package", "package", sanitizedPackageName, "error", err)
		os.Exit(1)
	}

	fmt.Println("\nSuccess! Package analyzed and cached.")
	fmt.Printf("Files stored in: %s/%s/\n", savePath, packageName)
	fmt.Println("- func.txt")
	fmt.Println("- const.txt")
	fmt.Println("- struct.txt")
	fmt.Println("- var.txt")
}
