package users

import "errors"

var (
	errNoAdminUsers     = errors.New("no admin users found in plist")
	errLastUserNotFound = errors.New("last user not found in plist")
)
