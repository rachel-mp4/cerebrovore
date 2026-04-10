package utils

import (
	"net/url"
	"strings"
)

func IsSafeLink(l string) bool {
	u, err := url.Parse(l)
	if err != nil {
		return false
	}
	if u.Scheme == "" || u.Host == "" {
		return false
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		return true
	default:
		return false
	}
}
