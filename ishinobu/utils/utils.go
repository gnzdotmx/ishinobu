package utils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func GetOutputFileName(moduleName, format, outputDir string) string {
	fileName := moduleName + "." + format
	return filepath.Join(outputDir, fileName)
}

// ListFiles lists all files that match the given glob-like pattern.
// Example pattern: /path/starts*/*ends/file-*.asl
func ListFiles(pattern string) ([]string, error) {
	// Clean the pattern and split into components
	pattern = filepath.Clean(pattern)
	parts := splitPath(pattern)

	// Determine the root directory
	var roots []string
	if filepath.IsAbs(pattern) {
		roots = []string{string(os.PathSeparator)}
	} else {
		roots = []string{"."}
	}

	// Iterate over each part of the pattern
	for _, part := range parts {
		var matches []string
		for _, root := range roots {
			matched, err := filepath.Glob(filepath.Join(root, part))
			if err != nil {
				return nil, err
			}
			matches = append(matches, matched...)
		}
		if len(matches) == 0 {
			return nil, errors.New("no matches found")
		}
		roots = matches
	}

	// Filter out directories; only return files
	var files []string
	for _, path := range roots {
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			files = append(files, path)
		}
	}

	return files, nil
}

// splitPath splits the path into its components, handling both Unix and Windows separators.
func splitPath(path string) []string {
	// Replace backslashes with forward slashes for consistency
	path = strings.ReplaceAll(path, string(os.PathSeparator), "/")
	parts := strings.Split(path, "/")
	// Remove empty parts (can occur with leading '/')
	var cleanParts []string
	for _, part := range parts {
		if part != "" {
			cleanParts = append(cleanParts, part)
		}
	}
	return cleanParts
}