// This module reads and parses:
// - SSH known_hosts files for each user on disk
// - SSH authorized_keys files for each user on disk
// Relevant fields:
// - src_name: Name of the source file (known_hosts or authorized_keys)
// - user: Username from the path
// - bits: Number of bits in the key
// - fingerprint: SSH key fingerprint
// - host: Hostname or IP address
// - keytype: Type of SSH key
package modules

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gnzdotmx/ishinobu/ishinobu/mod"
	"github.com/gnzdotmx/ishinobu/ishinobu/utils"
)

type SSHModule struct {
	Name        string
	Description string
}

func init() {
	module := &SSHModule{
		Name:        "ssh",
		Description: "Collects and parses SSH known_hosts and authorized_keys files"}
	mod.RegisterModule(module)
}

func (m *SSHModule) GetName() string {
	return m.Name
}

func (m *SSHModule) GetDescription() string {
	return m.Description
}

func (m *SSHModule) Run(params mod.ModuleParams) error {
	// Find all .ssh directories for users
	sshDirs, err := filepath.Glob("/Users/*/.ssh")
	if err != nil {
		params.Logger.Debug("Error listing SSH directories in /Users: %v", err)
	}

	// Also check /private/var for system users
	privateDirs, err := filepath.Glob("/private/var/*/.ssh")
	if err != nil {
		params.Logger.Debug("Error listing SSH directories in /private/var: %v", err)
	}

	// Combine both directory lists
	sshDirs = append(sshDirs, privateDirs...)

	outputFileName := utils.GetOutputFileName(m.GetName(), params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}
	defer writer.Close() // Make sure we close the writer when we're done

	for _, sshDir := range sshDirs {
		// Get username from path
		username := utils.GetUsernameFromPath(sshDir)

		// Check for known_hosts file
		knownHostsPath := filepath.Join(sshDir, "known_hosts")
		if _, err := os.Stat(knownHostsPath); err == nil {
			err = parseSSHFile(knownHostsPath, username, "known_hosts", params, writer)
			if err != nil {
				params.Logger.Debug("Error parsing known_hosts for %s: %v", username, err)
			}
		} else {
			params.Logger.Debug("known_hosts not found for user %s", username)
		}

		// Check for authorized_keys file
		authorizedKeysPath := filepath.Join(sshDir, "authorized_keys")
		if _, err := os.Stat(authorizedKeysPath); err == nil {
			err = parseSSHFile(authorizedKeysPath, username, "authorized_keys", params, writer)
			if err != nil {
				params.Logger.Debug("Error parsing authorized_keys for %s: %v", username, err)
			}
		} else {
			params.Logger.Debug("authorized_keys not found for user %s", username)
		}
	}

	return nil
}

func parseSSHFile(filePath, username, srcName string, params mod.ModuleParams, writer *utils.DataWriter) error {
	// Run ssh-keygen to get key information
	cmd := exec.Command("ssh-keygen", "-l", "-f", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "is not a public key file") {
			return fmt.Errorf("file is not a public key file: %v", err)
		}
		return fmt.Errorf("error running ssh-keygen: %v", err)
	}

	// Process each line of output
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse the ssh-keygen output
		parts := strings.Fields(line)
		if len(parts) < 4 {
			continue
		}

		recordData := make(map[string]interface{})
		recordData["src_name"] = srcName
		recordData["user"] = username
		recordData["bits"] = parts[0]
		recordData["fingerprint"] = parts[1]
		recordData["host"] = parts[2]
		recordData["keytype"] = strings.Trim(parts[3], "()")

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			Data:                recordData,
			SourceFile:          filePath,
		}

		err = writer.WriteRecord(record)
		if err != nil {
			params.Logger.Debug("Failed to write record: %v", err)
			continue
		}
	}

	return nil
}
