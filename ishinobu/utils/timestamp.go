package utils

import (
	"time"
)

const TimeFormat = "2006-01-02T15:04:05Z07:00"

func Now() string {
	return time.Now().Format(TimeFormat)
}

func ParseTimestamp(value string) (string, error) {
	location, _ := time.LoadLocation("UTC")
	t, err := time.ParseInLocation("2006-01-02 15:04:05.999999-0700", value, location)
	if err != nil {
		return "", err
	}

	return t.Format(TimeFormat), nil
}
