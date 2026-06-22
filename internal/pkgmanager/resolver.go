package pkgmanager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolvePackagePath converts a package string (possibly with version like import@v1.0.0)
// into an absolute file system path using the local Go module cache.
func ResolvePackagePath(packageName, localCacheDir string) (string, error) {
	if packageName == "" {
		return "", fmt.Errorf("package name cannot be empty")
	}

	// Expand tilde in localCacheDir if present
	expandedCacheDir := localCacheDir
	if strings.HasPrefix(localCacheDir, "~/") {
		home := os.Getenv("HOME")
		if home == "" {
			return "", fmt.Errorf("HOME environment variable not set")
		}
		expandedCacheDir = filepath.Join(home, localCacheDir[2:])
	}

	// If it already looks like an absolute path, return it
	if filepath.IsAbs(packageName) {
		// Ensure the absolute path is cleaned to prevent traversal
		return filepath.Clean(packageName), nil
	}

	// Handle the import@version format (e.g., git.../api/v3@v3.99.0)
	parts := strings.Split(packageName, "@")
	importPath := parts[0]
	version := ""
	if len(parts) > 1 {
		version = parts[1]
	} else {
		return "", fmt.Errorf("package version must be specified (e.g., package@v1.0.0)")
	}

	if version == "" {
		return "", fmt.Errorf("version is empty")
	}

	// Go module cache path: $GOPATH/pkg/mod/<module>/<version>
	// The importPath might be a package within a module, so we need to find the module root.
	fragments := strings.Split(importPath, "/")
	for i := len(fragments) - 1; i >= 0; i-- {
		currentPrefix := strings.Join(fragments[:i], "/")
		attemptPath := filepath.Join(expandedCacheDir, currentPrefix+"@"+version)

		// Security check: ensure attemptPath is still within expandedCacheDir
		if !strings.HasPrefix(attemptPath, expandedCacheDir) {
			continue
		}

		if _, err := os.Stat(attemptPath); err == nil { //nolint:gosec
			remainingPath := strings.Join(fragments[i:], "/")
			finalPath := filepath.Join(attemptPath, remainingPath)
			// Final safety check for path traversal
			if !strings.HasPrefix(finalPath, expandedCacheDir) {
				return "", fmt.Errorf("security error: resolved path is outside cache directory")
			}
			return finalPath, nil
		}
	}

	// Fallback: just try as is if it exists (for cases where user manually put @version at the end)
	fullPath := filepath.Join(expandedCacheDir, importPath+"@"+version)
	if !strings.HasPrefix(fullPath, expandedCacheDir) {
		return "", fmt.Errorf("security error: resolved path is outside cache directory")
	}

	if _, err := os.Stat(fullPath); err == nil { //nolint:gosec
		return fullPath, nil
	}

	return "", fmt.Errorf("package not found in local cache: %s", fullPath)
}
