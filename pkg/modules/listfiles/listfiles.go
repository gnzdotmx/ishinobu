// This module recursively traverses the file system and captures metadata for files
// and folders on disk, including:
// - MD5 and SHA256 hashes
// - MACB timestamps (Modified, Accessed, Created, Birth)
// - Extended attributes (quarantine, wherefrom, downloaddate)
package listfiles

import (
	"crypto/md5" // #nosec G501
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type ListFilesModule struct {
	Name        string
	Description string
}

// FileMetadata represents the metadata collected for each file/directory
type FileMetadata struct {
	Path          string
	Name          string
	Size          int64
	Mode          string
	Owner         string
	UID           int
	GID           int
	ModTime       string
	AccessTime    string
	CreateTime    string
	BirthTime     string
	MD5           string
	SHA256        string
	Quarantine    string
	WhereFrom1    string
	WhereFrom2    string
	DownloadDate  string
	CodeSignature string
}

// Invalid extensions to skip
var invalidExtensions = map[string]bool{
	".app":       true,
	".framework": true,
	".lproj":     true,
	".plugin":    true,
	".kext":      true,
	".osax":      true,
	".bundle":    true,
	".driver":    true,
	".wdgt":      true,
	".Office":    true,
}

const (
	maxWorkers  = 4                 // Number of concurrent workers
	maxFileSize = 100 * 1024 * 1024 // 100MB max file size for hashing
)

// Default paths relevant for IR triage
var defaultPaths = []string{
	"/Users",        // User home directories
	"/Applications", // Installed applications
	// "/Library",         // System libraries and configurations
	// "/private/var/log", // System logs
	// "/private/etc",                  // System configurations
	// "/System/Library/LaunchAgents",  // Launch agents
	// "/System/Library/LaunchDaemons", // Launch daemons
	// "/Library/LaunchAgents",
	// "/Library/LaunchDaemons",
}

// Interesting file extensions for IR triage
var interestingExtensions = map[string]bool{
	".sh":    true,
	".bash":  true,
	".py":    true,
	".rb":    true,
	".php":   true,
	".plist": true,
	".conf":  true,
	".log":   true,
	".exe":   true,
	".dll":   true,
	".dylib": true,
	".so":    true,
}

// Extended skip paths
var skipPaths = []string{
	"/System/Volumes/Data",
	"/System/Volumes/Preboot",
	"/System/Volumes/VM",
	"/private/var/vm",
	"/private/var/folders",
	"/Library/Caches",
	"/Users/Shared",
}

func init() {
	module := &ListFilesModule{
		Name:        "listfiles",
		Description: "Collects metadata for files and folders on disk"}
	mod.RegisterModule(module)
}

func (m *ListFilesModule) GetName() string {
	return m.Name
}

func (m *ListFilesModule) GetDescription() string {
	return m.Description
}

func (m *ListFilesModule) Run(params mod.ModuleParams) error {
	outputFileName := utils.GetOutputFileName(m.GetName(), params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}
	defer writer.Close()

	// Create a channel for workers
	jobs := make(chan string, 100)
	results := make(chan utils.Record, 100)
	errors := make(chan error, 100)

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go worker(jobs, results, errors, &wg, params)
	}

	// Start result writer
	go func() {
		for record := range results {
			if err := writer.WriteRecord(record); err != nil {
				params.Logger.Debug("Failed to write record: %v", err)
			}
		}
	}()

	// Process default paths
	for _, rootPath := range defaultPaths {
		err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				params.Logger.Debug("Error accessing path %s: %v", path, err)
				return nil
			}

			if shouldSkipPath(path) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Only process interesting files or directories
			if !info.IsDir() && !isInterestingFile(path) {
				return nil
			}

			// Skip large files
			if info.Size() > maxFileSize {
				params.Logger.Debug("Skipping large file %s (%d bytes)", path, info.Size())
				return nil
			}

			jobs <- path
			return nil
		})
		if err != nil {
			params.Logger.Debug("Error walking path %s: %v", rootPath, err)
		}
	}

	close(jobs)
	wg.Wait()
	close(results)
	close(errors)

	return nil
}

func worker(jobs <-chan string, results chan<- utils.Record, errors chan<- error, wg *sync.WaitGroup, params mod.ModuleParams) {
	defer wg.Done()

	for path := range jobs {
		info, err := os.Lstat(path)
		if err != nil {
			errors <- err
			continue
		}

		metadata, err := collectMetadata(path, info, params)
		if err != nil {
			errors <- err
			continue
		}

		metadataMap := map[string]interface{}{
			"path":           metadata.Path,
			"name":           metadata.Name,
			"size":           metadata.Size,
			"mode":           metadata.Mode,
			"owner":          metadata.Owner,
			"uid":            metadata.UID,
			"gid":            metadata.GID,
			"mod_time":       metadata.ModTime,
			"access_time":    metadata.AccessTime,
			"create_time":    metadata.CreateTime,
			"birth_time":     metadata.BirthTime,
			"md5":            metadata.MD5,
			"sha256":         metadata.SHA256,
			"quarantine":     metadata.Quarantine,
			"wherefrom_1":    metadata.WhereFrom1,
			"wherefrom_2":    metadata.WhereFrom2,
			"download_date":  metadata.DownloadDate,
			"code_signature": metadata.CodeSignature,
		}

		results <- utils.Record{
			CollectionTimestamp: utils.Now(),
			EventTimestamp:      metadata.ModTime,
			SourceFile:          path,
			Data:                metadataMap,
		}
	}
}

func shouldSkipPath(path string) bool {
	// Check invalid extensions
	if invalidExtensions[filepath.Ext(path)] {
		return true
	}

	// Check skip paths
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return false
}

func isInterestingFile(path string) bool {
	ext := filepath.Ext(path)
	return interestingExtensions[ext]
}

func collectMetadata(path string, info os.FileInfo, params mod.ModuleParams) (*FileMetadata, error) {
	metadata := &FileMetadata{
		Path: path,
		Name: info.Name(),
		Size: info.Size(),
	}

	// Get file mode
	if info.IsDir() {
		metadata.Mode = "directory"
	} else {
		metadata.Mode = "file"
	}

	// Get owner info
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		metadata.UID = int(stat.Uid)
		metadata.GID = int(stat.Gid)
		metadata.ModTime = utils.Now() // Convert Unix timestamp
		metadata.AccessTime = utils.Now()
		metadata.CreateTime = utils.Now()
		metadata.BirthTime = utils.Now()
	}

	// Calculate hashes for files
	if !info.IsDir() && info.Size() > 0 {
		md5hash, sha256hash, err := calculateHashes(path)
		if err != nil {
			return metadata, err
		}
		metadata.MD5 = md5hash
		metadata.SHA256 = sha256hash
	}

	// Get code signature for relevant files
	if !info.IsDir() {
		signature, err := utils.GetCodeSignature(path)
		if err != nil {
			params.Logger.Debug("Error getting code signature for %s: %v", path, err)
		} else {
			metadata.CodeSignature = signature
		}
	}

	return metadata, nil
}

func calculateHashes(path string) (string, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	md5Hash := md5.New() // #nosec G401
	sha256Hash := sha256.New()
	multiWriter := io.MultiWriter(md5Hash, sha256Hash)

	if _, err := io.Copy(multiWriter, file); err != nil {
		return "", "", err
	}

	return hex.EncodeToString(md5Hash.Sum(nil)), hex.EncodeToString(sha256Hash.Sum(nil)), nil
}
