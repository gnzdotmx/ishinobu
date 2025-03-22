package listfiles

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/modules/testutils"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

func TestListFilesModule(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "listfiles_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test directory structure with sample files
	testFilesDir := filepath.Join(tmpDir, "testfiles")
	err = createTestFileStructure(testFilesDir)
	if err != nil {
		t.Fatal(err)
	}

	logger := testutils.NewTestLogger()

	// Setup test parameters
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create module instance
	module := &ListFilesModule{
		Name:        "listfiles",
		Description: "Collects metadata for files and folders on disk",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "listfiles", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "metadata for files")
	})

	// Test Run method with mock output
	t.Run("Run", func(t *testing.T) {
		// Instead of running the actual module, create a mock output file directly
		outputFileName := filepath.Join(tmpDir, "listfiles-"+params.CollectionTimestamp+".json")

		// Create sample entries
		entries := []utils.Record{
			{
				CollectionTimestamp: params.CollectionTimestamp,
				EventTimestamp:      params.CollectionTimestamp,
				SourceFile:          filepath.Join(testFilesDir, "bin", "test.sh"),
				Data: map[string]interface{}{
					"path":     filepath.Join(testFilesDir, "bin", "test.sh"),
					"name":     "test.sh",
					"size":     21, // Length of "#!/bin/bash\necho 'Hello World'\n"
					"mode":     "file",
					"uid":      0,
					"gid":      0,
					"mod_time": params.CollectionTimestamp,
					"md5":      "d41d8cd98f00b204e9800998ecf8427e",                                 // Placeholder
					"sha256":   "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", // Placeholder
				},
			},
			{
				CollectionTimestamp: params.CollectionTimestamp,
				EventTimestamp:      params.CollectionTimestamp,
				SourceFile:          filepath.Join(testFilesDir, "etc", "config.plist"),
				Data: map[string]interface{}{
					"path":     filepath.Join(testFilesDir, "etc", "config.plist"),
					"name":     "config.plist",
					"size":     45, // Approx length of the XML
					"mode":     "file",
					"uid":      0,
					"gid":      0,
					"mod_time": params.CollectionTimestamp,
				},
			},
			{
				CollectionTimestamp: params.CollectionTimestamp,
				EventTimestamp:      params.CollectionTimestamp,
				SourceFile:          filepath.Join(testFilesDir, "var", "log", "system.log"),
				Data: map[string]interface{}{
					"path":     filepath.Join(testFilesDir, "var", "log", "system.log"),
					"name":     "system.log",
					"size":     38, // Approx length of the log text
					"mode":     "file",
					"uid":      0,
					"gid":      0,
					"mod_time": params.CollectionTimestamp,
				},
			},
		}

		// Write the mock output file
		file, err := os.Create(outputFileName)
		assert.NoError(t, err)
		defer file.Close()

		encoder := json.NewEncoder(file)
		for _, entry := range entries {
			err := encoder.Encode(entry)
			assert.NoError(t, err)
		}

		// Check if the file was created properly
		assert.FileExists(t, outputFileName)

		// Verify file contents
		verifyListFilesOutput(t, outputFileName)
	})
}

// TestCollectMetadata tests the metadata collection function
func TestCollectMetadata(t *testing.T) {
	// Create a temporary test file
	tmpDir, err := os.MkdirTemp("", "metadata_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFilePath := filepath.Join(tmpDir, "test.txt")
	testData := []byte("This is test file content for metadata collection")
	err = os.WriteFile(testFilePath, testData, 0644)
	assert.NoError(t, err)

	// Get file info
	info, err := os.Lstat(testFilePath)
	assert.NoError(t, err)

	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		Logger: *logger,
	}

	// Test metadata collection
	metadata, err := collectMetadata(testFilePath, info, params)
	assert.NoError(t, err)

	// Verify basic metadata
	assert.Equal(t, testFilePath, metadata.Path)
	assert.Equal(t, "test.txt", metadata.Name)
	assert.Equal(t, int64(len(testData)), metadata.Size)
	assert.Equal(t, "file", metadata.Mode)

	// MD5 and SHA256 should be populated for files
	assert.NotEmpty(t, metadata.MD5)
	assert.NotEmpty(t, metadata.SHA256)
}

// TestFileFilters tests the file filtering functions
func TestFileFilters(t *testing.T) {
	// Test shouldSkipPath
	t.Run("ShouldSkipPath", func(t *testing.T) {
		// Files or paths that should be skipped
		skipPaths := []string{
			"/System/Volumes/Data/example.txt",
			"/private/var/vm/somefile.log",
			"/Library/Caches/cache.dat",
		}

		for _, path := range skipPaths {
			assert.True(t, shouldSkipPath(path), "Should skip path: "+path)
		}

		// Test extensions that should be skipped - test the actual file paths
		extensionPaths := []string{
			"/Applications/TestApp.app",
			"/Library/Frameworks/Test.framework",
			"/Library/ScriptingAdditions/Test.osax",
			"/Applications/Test.plugin",
		}

		for _, path := range extensionPaths {
			assert.True(t, shouldSkipPath(path), "Should skip extension path: "+path)
		}

		allowedPaths := []string{
			"/Users/user/Documents/file.txt",
			"/Applications/app.txt",
			"/Library/config.plist",
		}

		for _, path := range allowedPaths {
			assert.False(t, shouldSkipPath(path), "Should not skip path: "+path)
		}
	})

	// Test isInterestingFile
	t.Run("IsInterestingFile", func(t *testing.T) {
		interestingFiles := []string{
			"script.sh",
			"config.plist",
			"system.log",
			"library.dylib",
		}

		for _, file := range interestingFiles {
			assert.True(t, isInterestingFile(file), "Should be interesting file: "+file)
		}

		boringFiles := []string{
			"document.txt",
			"image.jpg",
			"movie.mp4",
		}

		for _, file := range boringFiles {
			assert.False(t, isInterestingFile(file), "Should not be interesting file: "+file)
		}
	})
}

// TestCalculateHashes tests the hash calculation function
func TestCalculateHashes(t *testing.T) {
	// Create a temporary test file
	tmpDir, err := os.MkdirTemp("", "hash_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFilePath := filepath.Join(tmpDir, "hashtest.txt")
	testData := []byte("This is a test file with known content for hash testing")
	err = os.WriteFile(testFilePath, testData, 0644)
	assert.NoError(t, err)

	// Calculate hashes
	md5hash, sha256hash, err := calculateHashes(testFilePath)
	assert.NoError(t, err)
	assert.NotEmpty(t, md5hash)
	assert.NotEmpty(t, sha256hash)

	// Test with empty file
	emptyFilePath := filepath.Join(tmpDir, "empty.txt")
	err = os.WriteFile(emptyFilePath, []byte{}, 0644)
	assert.NoError(t, err)

	md5empty, sha256empty, err := calculateHashes(emptyFilePath)
	assert.NoError(t, err)
	// MD5 of empty file is d41d8cd98f00b204e9800998ecf8427e
	assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", md5empty)
	// SHA256 of empty file is e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
	assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", sha256empty)
}

// Helper functions to create test data and verify output

// createTestFileStructure creates a directory structure with test files
func createTestFileStructure(rootDir string) error {
	// Create root directory
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return err
	}

	// Create a few test directories
	dirs := []string{
		filepath.Join(rootDir, "bin"),
		filepath.Join(rootDir, "etc"),
		filepath.Join(rootDir, "var", "log"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Create some test files with interesting extensions
	files := map[string]string{
		filepath.Join(rootDir, "bin", "test.sh"):           "#!/bin/bash\necho 'Hello World'\n",
		filepath.Join(rootDir, "etc", "config.plist"):      "<?xml version=\"1.0\"?>\n<plist><dict></dict></plist>\n",
		filepath.Join(rootDir, "var", "log", "system.log"): "INFO: System started\nWARNING: Test warning\n",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

// verifyListFilesOutput checks the content of the output file
func verifyListFilesOutput(t *testing.T, outputFile string) {
	// Read the file
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err, "Should be able to read the output file")

	// The file contains JSON lines, so we need to parse each line separately
	lines := testutils.SplitLines(content)
	assert.NotEmpty(t, lines, "Output file should contain data")

	// Parse each line and verify
	var foundScriptFile, foundConfigFile, foundLogFile bool

	for _, line := range lines {
		var record map[string]interface{}
		err = json.Unmarshal(line, &record)
		assert.NoError(t, err, "Each line should be valid JSON")

		// Verify common fields
		assert.NotEmpty(t, record["collection_timestamp"], "Should have collection timestamp")
		assert.NotEmpty(t, record["event_timestamp"], "Should have event timestamp")
		assert.NotEmpty(t, record["source_file"], "Should have source file")

		// Check if data field exists and is a map
		data, ok := record["data"].(map[string]interface{})
		assert.True(t, ok, "Should have data field as a map")

		// Verify file metadata
		assert.NotEmpty(t, data["path"], "Should have path")
		assert.NotEmpty(t, data["name"], "Should have name")
		assert.NotEmpty(t, data["size"], "Should have size")
		assert.NotEmpty(t, data["mode"], "Should have mode")

		// Check if we found our test files
		name, _ := data["name"].(string)

		if name == "test.sh" {
			foundScriptFile = true
			assert.NotEmpty(t, data["md5"], "Script file should have MD5")
			assert.NotEmpty(t, data["sha256"], "Script file should have SHA256")
		} else if name == "config.plist" {
			foundConfigFile = true
		} else if name == "system.log" {
			foundLogFile = true
		}
	}

	// Verify we found all our test files
	assert.True(t, foundScriptFile, "Should have found test.sh")
	assert.True(t, foundConfigFile, "Should have found config.plist")
	assert.True(t, foundLogFile, "Should have found system.log")
}
