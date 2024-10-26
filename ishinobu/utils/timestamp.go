package utils

import (
	"fmt"
	"strconv"
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

func ConvertCFAbsoluteTimeToDate(cfTimeStr string) (string, error) {
	cfTime, err := strconv.ParseFloat(cfTimeStr, 64)
	if err != nil {
		return "", fmt.Errorf("failed to parse float from string '%s': %w", cfTimeStr, err)
	}

	const epochOffset = 978307200
	unixTimestamp := cfTime + epochOffset
	seconds := int64(unixTimestamp)
	nanoseconds := int64((unixTimestamp - float64(seconds)) * 1e9)
	t := time.Unix(seconds, nanoseconds).UTC()
	formattedDate := t.Format("2006-01-02 15:04:05")

	return formattedDate, nil
}
