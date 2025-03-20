package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetOutputFileName(t *testing.T) {
	tests := []struct {
		name       string
		moduleName string
		format     string
		outputDir  string
		want       string
	}{
		{
			name:       "basic test",
			moduleName: "test",
			format:     "json",
			outputDir:  "output",
			want:       filepath.Join("output", "test.json"),
		},
		{
			name:       "with empty output dir",
			moduleName: "module",
			format:     "csv",
			outputDir:  "",
			want:       "module.csv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetOutputFileName(tt.moduleName, tt.format, tt.outputDir)
			if got != tt.want {
				t.Errorf("GetOutputFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListFiles(t *testing.T) {
	// Create temporary test directory structure
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFiles := []string{
		"test1.txt",
		"test2.txt",
		"sub/test3.txt",
	}

	for _, file := range testFiles {
		path := filepath.Join(tmpDir, file)
		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(path, []byte("test"), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name    string
		pattern string
		want    int
		wantErr bool
	}{
		{
			name:    "match all txt files",
			pattern: filepath.Join(tmpDir, "*.txt"),
			want:    2,
			wantErr: false,
		},
		{
			name:    "match files in subdirectory",
			pattern: filepath.Join(tmpDir, "sub/*.txt"),
			want:    1,
			wantErr: false,
		},
		{
			name:    "no matches",
			pattern: filepath.Join(tmpDir, "*.nonexistent"),
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ListFiles(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.want {
				t.Errorf("ListFiles() got %v files, want %v", len(got), tt.want)
			}
		})
	}
}

func TestGetUsernameFromPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "Users path",
			path: "/Users/johndoe/Documents",
			want: "johndoe",
		},
		{
			name: "private var path",
			path: "/private/var/johndoe/data",
			want: "johndoe",
		},
		{
			name: "other path",
			path: "/other/path",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetUsernameFromPath(tt.path)
			if got != tt.want {
				t.Errorf("GetUsernameFromPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source file
	srcPath := filepath.Join(tmpDir, "source.txt")
	content := []byte("test content")
	err = os.WriteFile(srcPath, content, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test copying
	dstPath := filepath.Join(tmpDir, "destination.txt")
	err = CopyFile(srcPath, dstPath)
	if err != nil {
		t.Errorf("CopyFile() error = %v", err)
	}

	// Verify content
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(dstContent) != string(content) {
		t.Errorf("CopyFile() content mismatch, got %v, want %v", string(dstContent), string(content))
	}

	// Test error case
	err = CopyFile("nonexistent", dstPath)
	if err == nil {
		t.Error("CopyFile() expected error for nonexistent source file")
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		item  string
		want  bool
	}{
		{
			name:  "item exists",
			slice: []string{"a", "b", "c"},
			item:  "b",
			want:  true,
		},
		{
			name:  "item does not exist",
			slice: []string{"a", "b", "c"},
			item:  "d",
			want:  false,
		},
		{
			name:  "empty slice",
			slice: []string{},
			item:  "a",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Contains(tt.slice, tt.item)
			if got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNestedValue(t *testing.T) {
	testMap := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": "value",
			},
		},
		"direct": "value",
	}

	tests := []struct {
		name string
		keys []string
		want interface{}
	}{
		{
			name: "nested value exists",
			keys: []string{"level1", "level2", "level3"},
			want: "value",
		},
		{
			name: "direct value",
			keys: []string{"direct"},
			want: "value",
		},
		{
			name: "path doesn't exist",
			keys: []string{"nonexistent", "path"},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetNestedValue(testMap, tt.keys...)
			if got != tt.want {
				t.Errorf("GetNestedValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
