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

// Parse from Oct 26 19:34:13 to 2006-01-02T15:04:05Z07:00 format
func ConvertDateString(input string) (string, error) {
	currentYear := time.Now().Year()
	inputWithYear := fmt.Sprintf("%s %d", input, currentYear)

	t, err := time.Parse("Jan 2 15:04:05 2006", inputWithYear)
	if err != nil {
		return "", err
	}

	output := t.Format("2006-01-02T15:04:05Z07:00")
	return output, nil
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
	formattedDate := t.Format(TimeFormat)

	return formattedDate, nil
}

func ParseChromeTimestamp(microsecondsStr string) string {
	// Parse the string to int64
	microseconds, err := strconv.ParseInt(microsecondsStr, 10, 64)
	if err != nil || microseconds == 0 {
		return ""
	}

	// Convert microseconds to nanoseconds
	nanoseconds := microseconds * 1000

	// Offset between Windows epoch (1601-01-01) and Unix epoch (1970-01-01) in seconds
	windowsToUnixOffsetSeconds := int64(11644473600)

	// Subtract the offset to get Unix time in nanoseconds
	unixNano := nanoseconds - (windowsToUnixOffsetSeconds * int64(time.Second))

	// Create time.Time from Unix time in nanoseconds
	timestamp := time.Unix(0, unixNano).UTC()

	// Format the result to ISO 8601 with nanosecond precision and return
	return timestamp.Format(TimeFormat)
}
