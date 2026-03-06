package utils

import (
	"strconv"
)

// IDToA converts an id to a base 36 alphanumeric number
// for use in UI
func IDToA(id uint32) string {
	return strconv.FormatUint(uint64(id), 36)
}

// AToID converts a base 36 alphanumeric number to an id
func AToID(s string) (id uint32, err error) {
	id64, err := strconv.ParseUint(s, 36, 32)
	if err != nil {
		return 0, err
	}
	return uint32(id64), nil
}

func AToIDf(s string) uint32 {
	id, _ := AToID(s)
	return id
}
