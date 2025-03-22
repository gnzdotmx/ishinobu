package terminal

import "errors"

var (
	errInvalidPadding            = errors.New("invalid padding")
	errInvalidPaddingLength      = errors.New("invalid padding length")
	errInvalidCiperTextSize      = errors.New("ciphertext is not a multiple of block size")
	errDataIVSizeMissmatch       = errors.New("data too short to contain IV")
	errInvalidDataDataFileHeader = errors.New("invalid data.data file header")
	errDataDataNotFound          = errors.New("required file data.data not found")
	errWindowsPlistNotFound      = errors.New("required file windows.plist not found")
	errInvalidBlockLength        = errors.New("invalid block length")
)
