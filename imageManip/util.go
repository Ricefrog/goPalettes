package imageManip

import (
	"image/color"
	"log"
	"strconv"
)

// hexStr in format "#FFFFFF"
func HexToNRGBA(hexStr string) color.NRGBA {
	r, err := strconv.ParseInt(hexStr[1:3], 16, 64)
	if err != nil {
		log.Fatal(err)
	}

	g, err := strconv.ParseInt(hexStr[3:5], 16, 64)
	if err != nil {
		log.Fatal(err)
	}

	b, err := strconv.ParseInt(hexStr[5:7], 16, 64)
	if err != nil {
		log.Fatal(err)
	}

	return color.NRGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: 0xff,
	}
}
