package utils

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestCompressOutput(t *testing.T) {
	// Create a temporary directory with a shorter path
	tempDir := "/tmp/test_temp"
	err := os.Mkdir(tempDir, 0755)
	if err != nil && !os.IsExist(err) {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some test files
	testFiles := []string{"test1.txt", "test2.txt"}
	testContent := "test content"

	for _, fname := range testFiles {
		filePath := filepath.Join(tempDir, fname)
		err := os.WriteFile(filePath, []byte(testContent), 0600)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", fname, err)
		}
	}

	// Create output file path
	outputFile := "/tmp/output.tar.gz"

	// Test compression
	err = CompressOutput(tempDir, outputFile)
	if err != nil {
		t.Fatalf("CompressOutput failed: %v", err)
	}

	// Verify the compressed file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatal("Compressed file was not created")
	}

	// Verify the contents of the compressed file
	err = verifyCompressedContents(t, outputFile, testFiles, testContent)
	if err != nil {
		t.Fatalf("Failed to verify compressed contents: %v", err)
	}

	// Remove the output file
	os.Remove(outputFile)
}

func verifyCompressedContents(t *testing.T, archivePath string, expectedFiles []string, expectedContent string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	foundFiles := make(map[string]bool)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Read file content
		content := make([]byte, header.Size)
		if _, err = io.ReadFull(tr, content); err != nil {
			return err
		}

		// Verify content
		if string(content) != expectedContent {
			t.Errorf("Content mismatch for file %s. Got %s, want %s",
				header.Name, string(content), expectedContent)
		}

		foundFiles[header.Name] = true
	}

	// Verify all expected files were found
	for _, expectedFile := range expectedFiles {
		if !foundFiles[filepath.Base(expectedFile)] {
			t.Errorf("Expected file %s not found in archive", expectedFile)
		}
	}

	return nil
}
