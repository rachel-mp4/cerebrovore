package utils

import (
	"errors"
	"fmt"
	"strconv"
)

func AToColor(s string) (color uint32, err error) {
	if len(s) < 7 {
		return 0, errors.New("color too short")
	}
	c, err := strconv.ParseUint(s[1:7], 16, 32)
	if err != nil {
		return 0, err
	}
	return uint32(c), nil
}

func AToColorf(s string) uint32 {
	c, _ := AToColor(s)
	return c
}

func ColorToA(color uint32) string {
	return fmt.Sprintf("#%06s", strconv.FormatUint(uint64(color), 16))
}

func ColorToAp(color *uint32) string {
	if color == nil {
		return "#452faa"
	}
	return ColorToA(*color)
}

func ColorIsDark(color *uint32) string {
	var cx string
	if color == nil {
		cx = "#000000"
	} else {
		cx = ColorToA(*color)
	}
	r, _ := strconv.ParseUint(cx[1:3], 16, 64)
	g, _ := strconv.ParseUint(cx[3:5], 16, 64)
	b, _ := strconv.ParseUint(cx[5:], 16, 64)
	luminance := (0.299*float32(r) + 0.587*float32(g) + 0.114*float32(b)) / 255.
	if luminance > 0.5 {
		return "light"
	}
	return "dark"
}
