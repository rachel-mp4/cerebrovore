package utils

import (
	"strconv"
)

func IDToA(id uint32) string {
	return strconv.FormatUint(uint64(id), 36)
}

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
