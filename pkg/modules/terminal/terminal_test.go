package terminal

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/testutils"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

func TestIntToUInt32(t *testing.T) {
	t.Run("Negative", func(t *testing.T) {
		_, err := intToUInt32(-1)
		assert.Error(t, err)
	})
	t.Run("TooLarge", func(t *testing.T) {
		_, err := intToUInt32(int(^uint32(0)) + 1)
		assert.Error(t, err)
	})
	t.Run("Valid", func(t *testing.T) {
		val, err := intToUInt32(42)
		assert.NoError(t, err)
		assert.Equal(t, uint32(42), val)
	})
}

func TestDecryptBlockErrors(t *testing.T) {
	key := make([]byte, 16)

	// Too short for IV
	data := make([]byte, 8)
	_, err := decryptBlock(data, key)
	assert.ErrorIs(t, err, errDataIVSizeMissmatch)

	// Valid IV, ciphertext not a multiple of block size
	iv := make([]byte, 16)
	ciphertext := make([]byte, 17) // not a multiple of block size
	data = append([]byte{}, iv...)
	data = append(data, ciphertext...)
	_, err = decryptBlock(data, key)
	assert.ErrorIs(t, err, errInvalidCiperTextSize)

	// Valid IV and ciphertext, but invalid padding
	ciphertext = make([]byte, 16)
	for i := range ciphertext {
		ciphertext[i] = 1
	}
	data = append([]byte{}, iv...)
	data = append(data, ciphertext...)
	data[len(data)-1] = 0 // invalid padding
	_, err = decryptBlock(data, key)
	assert.ErrorIs(t, err, errInvalidPaddingLength)
}

func TestProcessTerminalStateErrors(t *testing.T) {
	tmpDir := t.TempDir()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *testutils.NewTestLogger(),
	}
	writer := &testutils.TestDataWriter{Records: []utils.Record{}}
	location := filepath.Join(tmpDir, "fakeuser")
	err := os.MkdirAll(location, 0755)
	require.NoError(t, err)

	t.Run("MissingWindowsPlist", func(t *testing.T) {
		err := os.WriteFile(filepath.Join(location, "data.data"), []byte("NSCR1000"), 0600)
		assert.NoError(t, err)
		err = processTerminalState(location, params, writer)
		assert.ErrorIs(t, err, errWindowsPlistNotFound)
	})
	t.Run("MissingDataData", func(t *testing.T) {
		plistContent := `<?xml version=\"1.0\" encoding=\"UTF-8\"?><plist version=\"1.0\"><dict></dict></plist>`
		err := os.WriteFile(filepath.Join(location, "windows.plist"), []byte(plistContent), 0600)
		assert.NoError(t, err)
		os.Remove(filepath.Join(location, "data.data"))
		err = processTerminalState(location, params, writer)
		assert.ErrorIs(t, err, errDataDataNotFound)
	})
	// Invalid header
	t.Run("InvalidHeader", func(t *testing.T) {
		plistContent := `<?xml version=\"1.0\" encoding=\"UTF-8\"?><plist version=\"1.0\"><dict></dict></plist>`
		err := os.WriteFile(filepath.Join(location, "windows.plist"), []byte(plistContent), 0600)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(location, "data.data"), []byte("BADHDR"), 0600)
		assert.NoError(t, err)
		err = processTerminalState(location, params, writer)
		assert.ErrorIs(t, err, errInvalidDataDataFileHeader)
	})
}

func TestCollectTerminalHistory(t *testing.T) {
	tmpDir := t.TempDir()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *testutils.NewTestLogger(),
	}
	module := &TerminalModule{Name: "terminal", Description: "desc"}

	// Create a fake history file
	histFile := filepath.Join(tmpDir, ".bash_history")
	err := os.WriteFile(histFile, []byte("ls\necho test\n"), 0600)
	assert.NoError(t, err)

	err = module.collectTerminalHistory([]string{histFile}, params)
	assert.NoError(t, err)
}

func TestCollectTerminalStateNoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *testutils.NewTestLogger(),
	}
	module := &TerminalModule{Name: "terminal", Description: "desc"}
	err := module.collectTerminalState([]string{}, params)
	assert.NoError(t, err)
}

func TestRunHandlesErrorsGracefully(t *testing.T) {
	tmpDir := t.TempDir()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *testutils.NewTestLogger(),
	}
	module := &TerminalModule{Name: "terminal", Description: "desc"}
	// Should not panic or return error even if no files exist
	err := module.Run(params)
	assert.NoError(t, err)
}

func TestGetDescription_Coverage(t *testing.T) {
	m := &TerminalModule{Description: "desc"}
	assert.Equal(t, "desc", m.GetDescription())
}

func TestCollectTerminalHistory_MultiUserAndError(t *testing.T) {
	tmpDir := t.TempDir()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *testutils.NewTestLogger(),
	}
	module := &TerminalModule{Name: "terminal", Description: "desc"}

	user1 := "user1"
	user2 := "user2"
	histFile1 := filepath.Join(tmpDir, user1+"_history")
	histFile2 := filepath.Join(tmpDir, user2+"_history")
	err := os.WriteFile(histFile1, []byte("ls\nwhoami\n"), 0600)
	assert.NoError(t, err)
	err = os.WriteFile(histFile2, []byte("pwd\n"), 0600)
	assert.NoError(t, err)

	badFile := filepath.Join(tmpDir, "bad_history")
	err = os.WriteFile(badFile, []byte("fail"), 0000)
	assert.NoError(t, err)

	err = module.collectTerminalHistory([]string{histFile1, histFile2, badFile}, params)
	assert.NoError(t, err)
}

func TestProcessTerminalState_BlockBranches(t *testing.T) {
	tmpDir := t.TempDir()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *testutils.NewTestLogger(),
	}
	writer := &testutils.TestDataWriter{Records: []utils.Record{}}
	location := filepath.Join(tmpDir, "user")
	err := os.MkdirAll(location, 0755)
	require.NoError(t, err)

	windowID := uint32(123)
	// key := []byte("1234567890abcdef") // removed unused variable
	plistContent := `<?xml version=\"1.0\" encoding=\"UTF-8\"?><plist version=\"1.0\"><dict></dict></plist>`
	err = os.WriteFile(filepath.Join(location, "windows.plist"), []byte(plistContent), 0600)
	assert.NoError(t, err)
	blockHeader := make([]byte, 8)
	binary.BigEndian.PutUint32(blockHeader[:4], windowID)
	binary.BigEndian.PutUint32(blockHeader[4:8], uint32(8))
	block := append([]byte{}, blockHeader...)
	block = append(block, []byte{}...)
	data := append([]byte("NSCR1000"), block...)
	err = os.WriteFile(filepath.Join(location, "data.data"), data, 0600)
	assert.NoError(t, err)
	// This test ensures the block loop is exercised with valid headers, but no records are written.
	err = processTerminalState(location, params, writer)
	assert.Error(t, err) // Should error due to plist parse failure
}

func TestGetDecryptionKey_Coverage(t *testing.T) {
	windowID := uint32(42)
	key := []byte("keybytes01234567")
	// Valid case
	windowsData := map[string]interface{}{
		"WindowList": []interface{}{
			map[string]interface{}{
				"StateID":   windowID,
				"NSDataKey": key,
			},
		},
	}
	k, ok := getDecryptionKey(windowsData, windowID)
	assert.True(t, ok)
	assert.Equal(t, key, k)
	// WindowList wrong type
	windowsData = map[string]interface{}{"WindowList": 123}
	_, ok = getDecryptionKey(windowsData, windowID)
	assert.False(t, ok)
	// Window missing fields
	windowsData = map[string]interface{}{"WindowList": []interface{}{map[string]interface{}{}}}
	_, ok = getDecryptionKey(windowsData, windowID)
	assert.False(t, ok)
	// StateID wrong type
	windowsData = map[string]interface{}{"WindowList": []interface{}{map[string]interface{}{"StateID": "notuint32"}}}
	_, ok = getDecryptionKey(windowsData, windowID)
	assert.False(t, ok)
}

func TestWriteTerminalStateRecord_Coverage(t *testing.T) {
	writer := &testutils.TestDataWriter{Records: []utils.Record{}}
	params := mod.ModuleParams{
		LogsDir:             "",
		OutputDir:           "",
		ExportFormat:        "json",
		CollectionTimestamp: "2024-01-01T00:00:00Z",
		Logger:              *testutils.NewTestLogger(),
	}
	// Full valid nested structure
	terminalState := map[string]interface{}{
		"NSTitle": "title",
		"TTWindowState": map[string]interface{}{
			"Window Settings": []interface{}{
				map[string]interface{}{
					"Tab Working Directory URL":        "file:///Users/test",
					"Tab Working Directory URL String": "/Users/test",
					"Tab Contents v2":                  []interface{}{"line1", "line2"},
				},
			},
		},
	}
	err := writeTerminalStateRecord(terminalState, "user", 1, 1, writer, params)
	assert.NoError(t, err)
	assert.Len(t, writer.Records, 2)
	// Missing TTWindowState
	err = writeTerminalStateRecord(map[string]interface{}{}, "user", 1, 1, writer, params)
	assert.NoError(t, err)
	// Window Settings wrong type
	terminalState = map[string]interface{}{
		"TTWindowState": map[string]interface{}{"Window Settings": 123},
	}
	err = writeTerminalStateRecord(terminalState, "user", 1, 1, writer, params)
	assert.NoError(t, err)
	// Tab Contents v2 wrong type
	terminalState = map[string]interface{}{
		"TTWindowState": map[string]interface{}{
			"Window Settings": []interface{}{
				map[string]interface{}{"Tab Contents v2": 123},
			},
		},
	}
	err = writeTerminalStateRecord(terminalState, "user", 1, 1, writer, params)
	assert.NoError(t, err)
}
