// Description: This module collects and parses Terminal.app saved state files and terminal histories.
// It collects the saved state files from the following paths:
// - /Users/*/Library/Saved Application State/com.apple.Terminal.savedState
// - /private/var/*/Library/Saved Application State/com.apple.Terminal.savedState
// It also collects the terminal histories from the following paths:
// - /Users/*/.*_history
// - /Users/*/.bash_sessions/*
// - /private/var/*/.*_history
package modules

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnzdotmx/ishinobu/ishinobu/mod"
	"github.com/gnzdotmx/ishinobu/ishinobu/utils"
)

type TerminalModule struct {
	Name        string
	Description string
}

func init() {
	module := &TerminalModule{
		Name:        "terminal",
		Description: "Collects and parses Terminal.app saved state files and terminal histories"}
	mod.RegisterModule(module)
}

func (m *TerminalModule) GetName() string {
	return m.Name
}

func (m *TerminalModule) GetDescription() string {
	return m.Description
}

func (m *TerminalModule) Run(params mod.ModuleParams) error {
	// Run Terminal.savedState collection
	if err := m.collectTerminalState(params); err != nil {
		params.Logger.Debug("Error collecting Terminal.savedState: %v", err)
	}

	// Run terminal history collection
	if err := m.collectTerminalHistory(params); err != nil {
		params.Logger.Debug("Error collecting terminal histories: %v", err)
	}

	return nil
}

func (m *TerminalModule) collectTerminalState(params mod.ModuleParams) error {
	locations, err := filepath.Glob("/Users/*/Library/Saved Application State/com.apple.Terminal.savedState")
	if err != nil {
		params.Logger.Debug("Error listing Terminal locations: %v", err)
		return err
	}

	// Also check private/var locations
	varLocations, err := filepath.Glob("/private/var/*/Library/Saved Application State/com.apple.Terminal.savedState")
	if err != nil {
		params.Logger.Debug("Error listing Terminal var locations: %v", err)
		return err
	}

	locations = append(locations, varLocations...)

	if len(locations) == 0 {
		params.Logger.Info("No Terminal.savedState files were found")
		return nil
	}

	// Create a map to store writers for each user
	writers := make(map[string]*utils.DataWriter)

	for _, location := range locations {
		username := utils.GetUsernameFromPath(location)

		// Create writer for this user if it doesn't exist
		if _, exists := writers[username]; !exists {
			outputFileName := utils.GetOutputFileName(m.GetName()+"-state", params.ExportFormat, params.OutputDir)
			writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
			if err != nil {
				params.Logger.Debug("Error creating writer for user %s: %v", username, err)
				continue
			}
			writers[username] = writer
		}

		err := processTerminalState(location, params, writers[username])
		if err != nil {
			params.Logger.Debug("Error processing Terminal state at %s: %v", location, err)
		}
	}

	return nil
}

func (m *TerminalModule) collectTerminalHistory(params mod.ModuleParams) error {
	paths := []string{"/Users/*/.*_history", "/Users/*/.bash_sessions/*",
		"/private/var/*/.*_history", "/private/var/*/.bash_sessions/*"}
	var expandedPaths []string
	for _, path := range paths {
		expandedPath, err := filepath.Glob(path)
		if err != nil {
			continue
		}
		expandedPaths = append(expandedPaths, expandedPath...)
	}

	// Create a map to store writers for each user
	writers := make(map[string]*utils.DataWriter)

	for _, path := range expandedPaths {
		username := utils.GetUsernameFromPath(path)

		// Create writer for this user if it doesn't exist
		if _, exists := writers[username]; !exists {
			outputFileName := utils.GetOutputFileName(m.GetName()+"-history", params.ExportFormat, params.OutputDir)
			writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
			if err != nil {
				params.Logger.Debug("Error creating writer for user %s: %v", username, err)
				continue
			}
			writers[username] = writer
		}

		file, err := os.Open(path)
		if err != nil {
			params.Logger.Debug("Error opening file: %v", err)
			continue
		}

		r := bufio.NewReader(file)
		for {
			line, _, err := r.ReadLine()

			if len(line) > 0 {
				recordData := make(map[string]interface{})
				recordData["username"] = username
				recordData["command"] = string(line)

				record := utils.Record{
					CollectionTimestamp: params.CollectionTimestamp,
					EventTimestamp:      params.CollectionTimestamp,
					Data:                recordData,
					SourceFile:          path,
				}

				err := writers[username].WriteRecord(record)
				if err != nil {
					params.Logger.Debug("Error writing record: %v", err)
				}
			}

			if err != nil {
				break
			}
		}
		file.Close()
	}

	// Close all writers
	for _, writer := range writers {
		writer.Close()
	}

	return nil
}

func processTerminalState(location string, params mod.ModuleParams, writer *utils.DataWriter) error {
	// Get username from path
	user := utils.GetUsernameFromPath(location)

	// Check for required files
	windowsPlist := filepath.Join(location, "windows.plist")
	dataFile := filepath.Join(location, "data.data")

	if _, err := os.Stat(windowsPlist); os.IsNotExist(err) {
		return fmt.Errorf("required file windows.plist not found for user %s", user)
	}

	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		return fmt.Errorf("required file data.data not found for user %s", user)
	}

	// Read data.data file
	data, err := os.ReadFile(dataFile)
	if err != nil {
		return fmt.Errorf("error reading data.data: %v", err)
	}

	// Check file header
	if string(data[:8]) != "NSCR1000" {
		return fmt.Errorf("invalid data.data file header")
	}

	// Parse windows.plist
	windowsPlistData, err := os.ReadFile(windowsPlist)
	if err != nil {
		return fmt.Errorf("error reading windows.plist: %v", err)
	}

	// Try to parse as binary plist first
	windowsData, err := utils.ParseBiPList(string(windowsPlistData))
	if err != nil {
		params.Logger.Debug("Failed to parse binary plist: %v", err)
		return err
	}

	// Process each NSCR1000 block
	blocks := strings.Split(string(data), "NSCR1000")
	for i, block := range blocks {
		if block == "" {
			continue
		}

		// Parse block header
		if len(block) < 8 {
			continue
		}

		windowID := binary.BigEndian.Uint32([]byte(block[:4]))
		blockSize := binary.BigEndian.Uint32([]byte(block[4:8]))

		if uint32(len(block))+8 != blockSize {
			params.Logger.Debug("Block size mismatch for window ID %d", windowID)
			continue
		}

		// Get decryption key from windows.plist
		key, ok := getDecryptionKey(windowsData, windowID)
		if !ok {
			params.Logger.Debug("No decryption key found for window ID %d", windowID)
			continue
		}

		// Process block data
		blockData := []byte(block[8:])
		decryptedData, err := decryptBlock(blockData, key)
		if err != nil {
			params.Logger.Debug("Failed to decrypt block %d: %v", i, err)
			continue
		}

		// Parse decrypted data as plist
		if idx := strings.Index(string(decryptedData), "bplist"); idx >= 0 {
			plistData := decryptedData[idx:]
			terminalState, err := utils.ParseBiPList(string(plistData))
			if err != nil {
				params.Logger.Debug("Failed to parse terminal state plist: %v", err)
				continue
			}

			// Write terminal state data
			err = writeTerminalStateRecord(terminalState, user, windowID, i+1, writer, params)
			if err != nil {
				params.Logger.Debug("Failed to write terminal state record: %v", err)
			}
		}
	}

	return nil
}

func getDecryptionKey(windowsData map[string]interface{}, windowID uint32) ([]byte, bool) {
	// Get the WindowList array from the plist data
	windowList, ok := windowsData["WindowList"].([]interface{})
	if !ok {
		return nil, false
	}

	// Search for the window with matching ID
	for _, window := range windowList {
		windowMap, ok := window.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if this is the window we're looking for
		if windowStateID, ok := windowMap["StateID"].(uint32); ok && windowStateID == windowID {
			// Get the NSDataKey which contains the decryption key
			if dataKey, ok := windowMap["NSDataKey"].([]byte); ok {
				return dataKey, true
			}
		}
	}

	return nil, false
}

func decryptBlock(data, key []byte) ([]byte, error) {
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %v", err)
	}

	// The IV (Initialization Vector) is typically the first block
	if len(data) < aes.BlockSize {
		return nil, fmt.Errorf("data too short to contain IV")
	}
	iv := data[:aes.BlockSize]
	ciphertext := data[aes.BlockSize:]

	// Create CBC decrypter
	mode := cipher.NewCBCDecrypter(block, iv)

	// Ciphertext must be a multiple of block size
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext is not a multiple of block size")
	}

	// Create output buffer and decrypt
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove PKCS#7 padding
	paddingLen := int(plaintext[len(plaintext)-1])
	if paddingLen > aes.BlockSize || paddingLen == 0 {
		return nil, fmt.Errorf("invalid padding length")
	}

	// Verify padding
	for i := len(plaintext) - paddingLen; i < len(plaintext); i++ {
		if plaintext[i] != byte(paddingLen) {
			return nil, fmt.Errorf("invalid padding")
		}
	}

	// Return decrypted data without padding
	return plaintext[:len(plaintext)-paddingLen], nil
}

func writeTerminalStateRecord(terminalState map[string]interface{}, user string, windowID uint32, blockIndex int, writer *utils.DataWriter, params mod.ModuleParams) error {
	recordData := make(map[string]interface{})
	recordData["user"] = user
	recordData["window_id"] = windowID
	recordData["datablock"] = blockIndex

	// Extract relevant fields from terminalState
	if ttWindow, ok := terminalState["TTWindowState"].(map[string]interface{}); ok {
		if settings, ok := ttWindow["Window Settings"].([]interface{}); ok && len(settings) > 0 {
			if setting, ok := settings[0].(map[string]interface{}); ok {
				recordData["window_title"] = terminalState["NSTitle"]
				recordData["tab_working_directory_url"] = setting["Tab Working Directory URL"]
				recordData["tab_working_directory_url_string"] = setting["Tab Working Directory URL String"]

				if contents, ok := setting["Tab Contents v2"].([]interface{}); ok {
					for i, line := range contents {
						if str, ok := line.(string); ok {
							recordData["line_index"] = i + 1
							recordData["line"] = strings.TrimSpace(str)

							record := utils.Record{
								CollectionTimestamp: params.CollectionTimestamp,
								EventTimestamp:      params.CollectionTimestamp,
								Data:                recordData,
								SourceFile:          fmt.Sprintf("Terminal.savedState/Window_%d", windowID),
							}

							if err := writer.WriteRecord(record); err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}

	return nil
}
