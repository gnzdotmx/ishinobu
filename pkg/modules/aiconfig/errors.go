package aiconfig

import "errors"

var (
	errConfigFileTooLarge = errors.New("config file exceeds size limit")
)
