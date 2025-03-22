package utils

import (
	"bytes"
	"fmt"

	"howett.net/plist"
)

func ParseBiPList(data string) (map[string]interface{}, error) {
	// Check for empty input
	if len(data) == 0 {
		return nil, errPlistEmpty
	}

	// Initialize a decoder from the string data
	decoder := plist.NewDecoder(bytes.NewReader([]byte(data)))

	// Decode the plist into a generic map
	var result map[string]interface{}
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding plist: %w", err)
	}

	return result, nil
}
