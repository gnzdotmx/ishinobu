package utils

import (
	"bytes"
	"os"
	"os/exec"
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
			return nil, errNoMatchesFound
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

func GetUsernameFromPath(path string) string {
	var user string
	if strings.Contains(path, "/Users/") {
		user = strings.Split(path, "/")[2]
	} else if strings.Contains(path, "/private/var/") {
		user = strings.Split(path, "/")[3]
	}

	return user
}

func CopyFile(src, dst string) error {
	// Read all content of src to data, may cause OOM for a large file.
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	// Write data to dst
	err = os.WriteFile(dst, data, 0600)
	if err != nil {
		return err
	}
	return nil
}

// GetCodeSignature returns the code signature information for a given program path.
// It uses the macOS `codesign` utility to verify the signature.
// Returns a string containing the signature information or an error if verification fails.
func GetCodeSignature(program string) (string, error) {
	if program == "" {
		return "", errEmptyProgramPath
	}

	// Run codesign -vv -d on the program
	cmd := exec.Command("codesign", "-vv", "-d", program)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Check if it's just an unsigned program
		if strings.Contains(stderr.String(), "code object is not signed") {
			return "Unsigned", nil
		}
		return "", err
	}

	// Parse and clean the output
	signature := out.String()
	if signature == "" {
		signature = stderr.String() // codesign sometimes writes to stderr even on success
	}

	// Clean up the signature output
	signature = strings.TrimSpace(signature)

	// If we got no useful output but the command succeeded,
	// try to get more detailed information
	if signature == "" {
		cmd = exec.Command("codesign", "--display", "--verbose=4", program)
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err == nil {
			signature = out.String()
			if signature == "" {
				signature = stderr.String()
			}
			signature = strings.TrimSpace(signature)
		}
	}

	if signature == "" {
		return "No signature information available", nil
	}

	return signature, nil
}

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper function to safely get nested map values
func GetNestedValue(m map[string]interface{}, keys ...string) interface{} {
	current := m
	for i, key := range keys {
		if i == len(keys)-1 {
			return current[key]
		}
		if val, ok := current[key].(map[string]interface{}); ok {
			current = val
		} else {
			return nil
		}
	}
	return nil
}
