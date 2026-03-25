package utils

import (
	"strconv"
)

// IDToA converts an id to a base 36 alphanumeric number
// for use in UI
func IDToA(id uint32) string {
	return strconv.FormatUint(uint64(id), 36)
}

func IntTo36A(i int) string {
	return strconv.FormatInt(int64(i), 36)
}

// AToID converts a base 36 alphanumeric number to an id
func AToID(s string) (id uint32, err error) {
	id64, err := strconv.ParseUint(s, 36, 32)
	if err != nil {
		return 0, err
	}
	return uint32(id64), nil
}

// AToEx converts a base
func AToEx(s string) (id uint32, iderr error, ex uint64, exerr error) {
	ex, exerr = strconv.ParseUint(s, 36, 64)
	id64, err := strconv.ParseUint(s, 36, 32)
	id = uint32(id64)
	iderr = err
	return
}

func AToIDf(s string) uint32 {
	id, _ := AToID(s)
	return id
}

// use this only for things we want to be correct at compile time, panics
// on error (basically i want to precompute commands, and this lets me
// know if the command's name is too long)
func AToIDp(s string) uint32 {
	id, err := AToID(s)
	if err != nil {
		panic(err)
	}
	return id
}

func AToExf(s string) uint64 {
	_, _, ex, _ := AToEx(s)
	return ex
}

func AToExp(s string) uint64 {
	_, _, ex, err := AToEx(s)
	if err != nil {
		panic(err)
	}
	return ex
}
