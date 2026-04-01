package utils

import (
	"time"
)

func RemainingTime(t time.Time) string {
	dur := t.Sub(time.Now())
	return dur.Round(time.Second).String()
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
