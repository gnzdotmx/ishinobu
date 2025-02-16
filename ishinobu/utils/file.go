package utils

import "os"

func ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}
