package main

import (
	"fmt"
	"math"
)

func main() {
	i := 1
	var v = generateVal(i)
	for v > 1 {
		fmt.Println(generateCss(i, v))
		i++
		v = generateVal(i)
	}

}

func generateVal(i int) float64 {
	offset := i - 22
	scale := float64(offset) * 0.25
	ataned := math.Atan(scale)
	rescaled := ataned / -2.6
	reoffset := rescaled + 1.5
	return math.Max(math.Min(2, reoffset), 1)
}

func generateCss(i int, f float64) string {
	return fmt.Sprintf(".transmission:nth-last-of-type(%d) {font-size:%frem;}", i, f)
}
