package main

import (
	"fmt"
	"math"
)

// generates css styles that you don't wanna do by hand: "go run ./cssgen > ./static/css/postsizes.go"
func main() {
	fmt.Println("@media (min-width: 36rem) {")
	i := 1
	var v = generateVal(i)
	for v > 1 {
		fmt.Println(generateCss(i, v))
		i++
		v = generateVal(i)
	}
	fmt.Println("}")
	for i := range 8 {
		fmt.Println(generateDelays(i))
		fmt.Println(generateDelays2(i))
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
	return fmt.Sprintf("#eats-ur-brain .tx:nth-last-of-type(%d) {font-size:%frem;}", i, f)
}

func generateDelays(i int) string {
	return fmt.Sprintf(".thread-tail:hover > div:nth-of-type(%d) {transition-delay: .%03ds; transition-duration: .%02ds}", i+1, i*11, 7+2*i)
}

func generateDelays2(i int) string {
	return fmt.Sprintf(".thread-tail:hover > div:nth-of-type(%d) .time {transition-delay: .%03ds; transition-duration: .%02ds}", i+1, i*11, 7+2*i)
}
