package utils

import "errors"

var (
	errPlistEmpty                  = errors.New("empty plist data")
	errNoMatchesFound              = errors.New("no matches found")
	errEmptyProgramPath            = errors.New("empty program path")
	errUnsupportedDataWriterFormat = errors.New("unsupported data writer format")
	errWriterInvalidRecordData     = errors.New("record data should be a map")
)
