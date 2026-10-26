package utils

import (
	"fmt"
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

func NiceParseDuration(s string) (d time.Duration, err error) {
	d, err = time.ParseDuration(s)
	if err == nil {
		return
	}
	var h, m, sec int
	_, err = fmt.Sscanf(s, "%d:%d:%d", &h, &m, &sec)
	if err != nil {
		h = 0
		_, err = fmt.Sscanf(s, "%d:%d", &m, &sec)
		if err != nil {
			m = 0
			_, err = fmt.Sscanf(s, "%d", &sec)
			if err != nil {
				return
			}
		}
	}
	return (time.Duration(h) * time.Hour) + (time.Duration(m) * time.Minute) + (time.Duration(sec) * time.Second), nil
}
