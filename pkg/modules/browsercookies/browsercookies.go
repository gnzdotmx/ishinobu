// This module collects and parses browser cookies from Chrome and Firefox.
// It collects the following information:
// - Chrome cookies
// - Firefox cookies
package browsercookies

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/modules/chrome"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type BrowserCookiesModule struct {
	Name        string
	Description string
}

func init() {
	module := &BrowserCookiesModule{
		Name:        "browsercookies",
		Description: "Collects and parses browser cookies from Chrome and Firefox"}
	mod.RegisterModule(module)
}

func (m *BrowserCookiesModule) GetName() string {
	return m.Name
}

func (m *BrowserCookiesModule) GetDescription() string {
	return m.Description
}

func (m *BrowserCookiesModule) Run(params mod.ModuleParams) error {
	// Chrome cookies
	chromeLocations, err := filepath.Glob("/Users/*/Library/Application Support/Google/Chrome")
	if err != nil {
		params.Logger.Debug("Error listing Chrome locations: %v", err)
		return err
	}

	for _, location := range chromeLocations {
		profilesDir, err := chrome.ChromeProfiles(location, m.GetName(), params)
		if err != nil {
			params.Logger.Debug("Error when collecting Chrome profiles: %v", err)
			continue
		}

		for _, profile := range profilesDir {
			err = collectChromeCookies(location, profile, m.GetName(), params)
			if err != nil {
				params.Logger.Debug("Error collecting Chrome cookies: %v", err)
			}
		}
	}

	// Firefox cookies
	firefoxLocations, err := filepath.Glob("/Users/*/Library/Application Support/Firefox/Profiles/*")
	if err != nil {
		params.Logger.Debug("Error listing Firefox locations: %v", err)
		return err
	}

	for _, location := range firefoxLocations {
		err = collectFirefoxCookies(location, m.GetName(), params)
		if err != nil {
			params.Logger.Debug("Error collecting Firefox cookies: %v", err)
		}
	}

	return nil
}

func collectChromeCookies(location, profileUsr, moduleName string, params mod.ModuleParams) error {
	// Create temporary directory
	ishinobuDir := "/tmp/ishinobu-Chrome-Cookies"
	if err := os.MkdirAll(ishinobuDir, os.ModePerm); err != nil {
		params.Logger.Debug("Failed to create directory /tmp/ishinobu-Chrome-Cookies", err)
		return err
	}

	// Extract user from location path
	user := strings.Split(location, "/")[2]
	outputFileName := utils.GetOutputFileName(moduleName+"-chrome-"+user+"-"+profileUsr, params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	cookiesDB := filepath.Join(location, profileUsr, "Cookies")
	userProfile := strings.Split(cookiesDB, "/")[len(strings.Split(cookiesDB, "/"))-1]
	dst := "/tmp/ishinobu-Chrome-Cookies/" + userProfile + "_chrome_cookies"
	err = utils.CopyFile(cookiesDB, dst)
	if err != nil {
		return fmt.Errorf("error copying file: %v", err)
	}

	query := `
		SELECT host_key, name, value, path, creation_utc, expires_utc, 
		last_access_utc, is_secure, is_httponly, has_expires, is_persistent,
		priority, encrypted_value, samesite, source_scheme
		FROM cookies
	`
	rows, err := utils.QuerySQLite(dst, query)
	if err != nil {
		return fmt.Errorf("error querying SQLite: %v", err)
	}

	recordData := make(map[string]interface{})
	for rows.Next() {
		var hostKey, name, value, path string
		var creationUtc, expiresUtc, lastAccessUtc, isSecure, isHttpOnly, hasExpires, isPersistent string
		var priority, encryptedValue, sameSite, sourceScheme string

		err := rows.Scan(&hostKey, &name, &value, &path, &creationUtc, &expiresUtc,
			&lastAccessUtc, &isSecure, &isHttpOnly, &hasExpires, &isPersistent,
			&priority, &encryptedValue, &sameSite, &sourceScheme)
		if err != nil {
			params.Logger.Debug("Error scanning row: %v", err)
			continue
		}

		recordData["chrome_profile"] = profileUsr
		recordData["host_key"] = hostKey
		recordData["name"] = name
		recordData["value"] = value
		recordData["path"] = path
		recordData["creation_utc"] = utils.ParseChromeTimestamp(creationUtc)
		recordData["expires_utc"] = utils.ParseChromeTimestamp(expiresUtc)
		recordData["last_access_utc"] = utils.ParseChromeTimestamp(lastAccessUtc)
		recordData["is_secure"] = isSecure
		recordData["is_httponly"] = isHttpOnly
		recordData["has_expires"] = hasExpires
		recordData["is_persistent"] = isPersistent
		recordData["priority"] = priority
		recordData["encrypted_value"] = len(encryptedValue) > 0
		recordData["samesite"] = sameSite
		recordData["source_scheme"] = sourceScheme

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      recordData["creation_utc"].(string),
			Data:                recordData,
			SourceFile:          cookiesDB,
		}

		err = writer.WriteRecord(record)
		if err != nil {
			params.Logger.Debug("Failed to write record: %v", err)
		}
	}

	// Cleanup
	err = os.RemoveAll(ishinobuDir)
	if err != nil {
		return fmt.Errorf("error removing directory /tmp/ishinobu-Chrome-Cookies: %v", err)
	}

	return nil
}

func collectFirefoxCookies(location, moduleName string, params mod.ModuleParams) error {
	// Create temporary directory
	ishinobuDir := "/tmp/ishinobu-Firefox-Cookies"
	if err := os.MkdirAll(ishinobuDir, os.ModePerm); err != nil {
		params.Logger.Debug("Failed to create directory /tmp/ishinobu-Firefox-Cookies", err)
		return err
	}

	profile := filepath.Base(location)
	user := strings.Split(location, "/")[2]

	outputFileName := utils.GetOutputFileName(moduleName+"-firefox-"+user+"-"+profile, params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	cookiesDB := filepath.Join(location, "cookies.sqlite")
	dst := "/tmp/ishinobu-Firefox-Cookies/" + profile + "_firefox_cookies"
	err = utils.CopyFile(cookiesDB, dst)
	if err != nil {
		return fmt.Errorf("error copying file: %v", err)
	}

	query := `
		SELECT host, name, value, path, creationTime, expiry, lastAccessed,
		isSecure, isHttpOnly, inBrowserElement, sameSite
		FROM moz_cookies
	`
	rows, err := utils.QuerySQLite(dst, query)
	if err != nil {
		return fmt.Errorf("error querying SQLite: %v", err)
	}

	recordData := make(map[string]interface{})
	for rows.Next() {
		var host, name, value, path string
		var creationTime, expiry, lastAccessed, isSecure, isHttpOnly, inBrowserElement, sameSite string

		err := rows.Scan(&host, &name, &value, &path, &creationTime, &expiry,
			&lastAccessed, &isSecure, &isHttpOnly, &inBrowserElement, &sameSite)
		if err != nil {
			params.Logger.Debug("Error scanning row: %v", err)
			continue
		}

		recordData["user"] = user
		recordData["profile"] = profile
		recordData["host"] = host
		recordData["name"] = name
		recordData["value"] = value
		recordData["path"] = path
		recordData["creation_time"] = utils.ParseChromeTimestamp(creationTime)
		recordData["expiry"] = expiry
		recordData["last_accessed"] = utils.ParseChromeTimestamp(lastAccessed)
		recordData["is_secure"] = isSecure
		recordData["is_httponly"] = isHttpOnly
		recordData["in_browser_element"] = inBrowserElement
		recordData["same_site"] = sameSite

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      recordData["creation_time"].(string),
			Data:                recordData,
			SourceFile:          cookiesDB,
		}

		err = writer.WriteRecord(record)
		if err != nil {
			params.Logger.Debug("Failed to write record: %v", err)
		}
	}

	// Cleanup
	err = os.RemoveAll(ishinobuDir)
	if err != nil {
		return fmt.Errorf("error removing directory /tmp/ishinobu-Firefox-Cookies: %v", err)
	}

	return nil
}
