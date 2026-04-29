package utils

import (
	"strings"
	"time"
)

func RemainingTime(t time.Time) string {
	dur := -time.Since(t)
	return dur.Round(time.Second).String()
}

func TimeSince(t *time.Time) string {
	if t == nil {
		return ""
	}
	dur := time.Since(*t)
	res := strings.TrimSuffix(dur.Round(time.Minute).String(), "0s")
	if res == "" {
		return "0m"
	}
	return res
}

func FormatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func FTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func ParseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
