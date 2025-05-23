package chrome

import "errors"

var (
	errNoProfileData     = errors.New("no profile data found")
	errNoContentSettings = errors.New("no content settings data found")
	errNoExceptions      = errors.New("no exceptions data found")
	errNoPopups          = errors.New("no popups data found")
	errNoExtensionName   = errors.New("no name found for extension")
)
