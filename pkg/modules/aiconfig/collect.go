package aiconfig

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

const maxConfigReadSize = 2 * 1024 * 1024 // 2MB

type fileInfo struct {
	size    int64
	modTime string
	sha256  string
}

func extractUsername(path string) string {
	parts := strings.Split(filepath.Clean(path), string(os.PathSeparator))
	for i, part := range parts {
		if part == "Users" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func extractProjectPath(configPath string, marker string) string {
	clean := filepath.Clean(configPath)
	idx := strings.Index(clean, marker)
	if idx <= 0 {
		return ""
	}
	return clean[:idx]
}

func statFile(path string) (fileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return fileInfo{}, err
	}

	fi := fileInfo{
		size:    info.Size(),
		modTime: info.ModTime().UTC().Format(utils.TimeFormat),
	}

	hash, err := hashFile(path, info.Size())
	if err == nil {
		fi.sha256 = hash
	}

	return fi, nil
}

func hashFile(path string, size int64) (string, error) {
	if size > maxConfigReadSize {
		return "", nil
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func readConfigFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > maxConfigReadSize {
		return nil, fmt.Errorf("%w: %s", errConfigFileTooLarge, path)
	}

	return os.ReadFile(path)
}

func envKeys(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	return keys
}

func riskIndicators(command string, args []string) []string {
	combined := strings.ToLower(command + " " + strings.Join(args, " "))
	indicators := []struct {
		substr string
		label  string
	}{
		{"npx ", "npx_invocation"},
		{"npm exec", "npm_exec"},
		{"curl ", "curl"},
		{"wget ", "wget"},
		{"bash -c", "bash_inline"},
		{"sh -c", "shell_inline"},
		{"| sh", "pipe_to_shell"},
		{"| bash", "pipe_to_bash"},
		{"base64", "base64"},
		{"eval ", "eval"},
		{"/tmp/", "tmp_path"},
	}

	var found []string
	for _, ind := range indicators {
		if strings.Contains(combined, ind.substr) {
			found = append(found, ind.label)
		}
	}
	return found
}

func writeRecord(writer utils.DataWriter, params mod.ModuleParams, sourceFile string, data map[string]interface{}) error {
	if fi, err := statFile(sourceFile); err == nil {
		data["file_size"] = fi.size
		data["file_mtime"] = fi.modTime
		if fi.sha256 != "" {
			data["sha256"] = fi.sha256
		}
	}

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      utils.Now(),
		SourceFile:          sourceFile,
		Data:                data,
	}

	return writer.WriteRecord(record)
}

func globPaths(patterns []string) []string {
	seen := make(map[string]struct{})
	var paths []string

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, match := range matches {
			if _, ok := seen[match]; ok {
				continue
			}
			seen[match] = struct{}{}
			paths = append(paths, match)
		}
	}

	return paths
}

func isReadableFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func globRuleFiles(patterns []string) []string {
	var files []string
	seen := make(map[string]struct{})

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue
			}
			if _, ok := seen[match]; ok {
				continue
			}
			seen[match] = struct{}{}
			files = append(files, match)
		}
	}

	return files
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
