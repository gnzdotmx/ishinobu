// This module enumerates current and deleted user profiles, identifies admin users and last logged in user.
// It collects the following information:
// - Deleted users
// - Admin users
// - Last logged in user
// - Current users
package users

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type UsersModule struct {
	Name        string
	Description string
}

func init() {
	module := &UsersModule{
		Name:        "users",
		Description: "Enumerates current and deleted user profiles, identifies admin users and last logged in user"}
	mod.RegisterModule(module)
}

func (m *UsersModule) GetName() string {
	return m.Name
}

func (m *UsersModule) GetDescription() string {
	return m.Description
}

func (m *UsersModule) Run(params mod.ModuleParams) error {
	outputFileName := utils.GetOutputFileName(m.GetName(), params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	// Get deleted users from preferences
	plistPath := "/Library/Preferences/com.apple.preferences.accounts.plist"
	deletedUsers, err := getDeletedUsers(plistPath)
	if err != nil {
		params.Logger.Debug("Error getting deleted users: %v", err)
	}

	// Write deleted users records
	for _, user := range deletedUsers {
		recordData := make(map[string]interface{})
		recordData["date_deleted"] = user.DateDeleted
		recordData["uniq_id"] = user.UniqueID
		recordData["user"] = user.Name
		recordData["real_name"] = user.RealName
		recordData["admin"] = ""
		recordData["lastloggedin_user"] = ""

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      user.DateDeleted,
			Data:                recordData,
			SourceFile:          "/Library/Preferences/com.apple.preferences.accounts.plist",
		}

		err = writer.WriteRecord(record)
		if err != nil {
			params.Logger.Debug("Failed to write record: %v", err)
		}
	}

	// Get admin users
	adminPlistPath := "/private/var/db/dslocal/nodes/Default/groups/admin.plist"
	adminUsers, err := getAdminUsers(adminPlistPath)
	if err != nil {
		params.Logger.Debug("Error getting admin users: %v", err)
	}

	// Get last logged in user
	loginPlistPath := "/Library/Preferences/com.apple.loginwindow.plist"
	lastUser, err := getLastLoggedInUser(loginPlistPath)
	if err != nil {
		params.Logger.Debug("Error getting last logged in user: %v", err)
	}

	// Get current users from /Users directory
	userDirs, err := filepath.Glob("/Users/*")
	if err != nil {
		params.Logger.Debug("Error listing user directories: %v", err)
	}

	// Filter out system directories
	systemDirs := map[string]bool{
		".localized": true,
		"Shared":     true,
		"Guest":      true,
	}

	for _, userDir := range userDirs {
		userName := filepath.Base(userDir)
		if systemDirs[userName] {
			continue
		}

		// Get user info from dscl
		// This would normally use dscl, but for forensics we'll read from the user plist
		userPlistPath := fmt.Sprintf("/private/var/db/dslocal/nodes/Default/users/%s.plist", userName)
		userInfo, err := getUserInfo(userPlistPath)
		if err != nil {
			params.Logger.Debug("Error getting user info for %s: %v", userName, err)
			continue
		}

		recordData := make(map[string]interface{})
		recordData["user"] = userName
		recordData["real_name"] = userInfo.RealName
		recordData["uniq_id"] = userInfo.UniqueID
		recordData["admin"] = utils.Contains(adminUsers, userName)
		recordData["lastloggedin_user"] = (userName == lastUser)

		// Get file timestamps
		fileInfo, err := os.Stat(userDir)
		if err == nil {
			switch stat := fileInfo.Sys().(type) {
			case *syscall.Stat_t:
				recordData["mtime"] = time.Unix(stat.Mtimespec.Sec, stat.Mtimespec.Nsec).UTC().Format(utils.TimeFormat)
				recordData["atime"] = time.Unix(stat.Atimespec.Sec, stat.Atimespec.Nsec).UTC().Format(utils.TimeFormat)
				recordData["ctime"] = time.Unix(stat.Ctimespec.Sec, stat.Ctimespec.Nsec).UTC().Format(utils.TimeFormat)
			default:
				params.Logger.Debug("Invalid file info type: %T", fileInfo.Sys())
			}
		}

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			Data:                recordData,
			SourceFile:          userDir,
		}

		err = writer.WriteRecord(record)
		if err != nil {
			params.Logger.Debug("Failed to write record: %v", err)
		}
	}

	return nil
}

type DeletedUser struct {
	DateDeleted string
	UniqueID    string
	Name        string
	RealName    string
}

func getDeletedUsers(plistPath string) ([]DeletedUser, error) {
	data, err := os.ReadFile(plistPath)
	if err != nil {
		return nil, err
	}

	plistData, err := utils.ParseBiPList(string(data))
	if err != nil {
		return nil, err
	}

	var deletedUsers []DeletedUser
	if deletedUsersData, ok := plistData["deletedUsers"].([]interface{}); ok {
		for _, userData := range deletedUsersData {
			if userMap, ok := userData.(map[string]interface{}); ok {

				date := userMap["date"].(string)
				uniqueID := userMap["dsAttrTypeStandard:UniqueID"].(string)
				realName := userMap["dsAttrTypeStandard:RealName"].(string)
				name := userMap["name"].(string)

				user := DeletedUser{
					DateDeleted: date,
					UniqueID:    uniqueID,
					Name:        name,
					RealName:    realName,
				}
				deletedUsers = append(deletedUsers, user)
			}
		}
	}

	return deletedUsers, nil
}

func getAdminUsers(adminPlistPath string) ([]string, error) {
	data, err := os.ReadFile(adminPlistPath)
	if err != nil {
		return nil, err
	}

	plistData, err := utils.ParseBiPList(string(data))
	if err != nil {
		return nil, err
	}

	if users, ok := plistData["users"].([]interface{}); ok {
		adminUsers := make([]string, len(users))
		for i, user := range users {
			adminUsers[i] = user.(string)
		}
		return adminUsers, nil
	}

	return nil, errNoAdminUsers
}

func getLastLoggedInUser(loginPlistPath string) (string, error) {
	data, err := os.ReadFile(loginPlistPath)
	if err != nil {
		return "", err
	}

	plistData, err := utils.ParseBiPList(string(data))
	if err != nil {
		return "", err
	}

	if lastUser, ok := plistData["lastUserName"].(string); ok {
		return lastUser, nil
	}

	return "", errLastUserNotFound
}

type UserInfo struct {
	UniqueID string
	RealName string
}

func getUserInfo(userPlistPath string) (*UserInfo, error) {
	data, err := os.ReadFile(userPlistPath)
	if err != nil {
		return nil, err
	}

	plistData, err := utils.ParseBiPList(string(data))
	if err != nil {
		return nil, err
	}

	userInfo := &UserInfo{}
	if uid, ok := plistData["uid"].([]interface{}); ok && len(uid) > 0 {
		userInfo.UniqueID = uid[0].(string)
	}
	if realname, ok := plistData["realname"].([]interface{}); ok && len(realname) > 0 {
		userInfo.RealName = realname[0].(string)
	}

	return userInfo, nil
}
