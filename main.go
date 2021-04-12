package main

import (
	"fmt"
	_ "image"
	_ "image/png"
	_ "image/color"
	"os"
	"errors"
	"log"
)

var IMAGE_PATH = "./images/"
var CORES_TO_USE = 4

type Domain struct {
	x int
	y int
}

func parseArgs() (s string, e string, err error) {
	args := os.Args[1:]
	if len(args) < 2 {
		err = errors.New("Not enough arguments.")
		return
	}

	s = args[0]
	e = args[1]
	return
}

func createColorFrequencyMap(img *Image) map[string]int {
	colFreqMap := make(map[string]int)
	domains := createDomains(img)
	for dom := range domains {
		go countColors(img, dom, &colFreqMap)
	}
}

// Split the bounds of the image into equal parts based on the
// number of cores being utilized.
func createDomains(img *Image) domains []Domain {
	xChunk := img.Bounds.Max.X / CORES_TO_USE
	maxY := img.Bounds.Max.Y

	for i := 1; i < CORES_TO_USE; i++ {
		newDom := Domain{
			xUpper: xChunk*i,
			yUpper: maxY,
			xLower: xChunk*i - xChunk,
			yLower: 0,
		}
		domains = append(domains, newDom)
	}

	// Last domain picks up remainder pixels.
	lastDom := Domain{
		xUpper: img.Bounds.Max.X,
		yUpper: maxY,
		xLower: (xChunk*(CORES_TO_USE-1)),
		yLower: 0,
	}
	domains = append(domains, newDom)

	return
}

func countColors(img *Image, dom Domain, colFreqMap *map[string]int) {
}

func main() {
	s, e, err := parseArgs()
	startPath := IMAGE_PATH+s
	endPath := IMAGE_PATH+e
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("startPath: %s, endPath: %s\n", startPath, endPath)

	originalFile, err := os.Open(startPath)
	if err != nil {
		log.Fatal(err)
	}
	defer originalFile.Close()

	originalData, _, err := image.Decode(originalFile)
	colorFrequencyMap := createColorFrequencyMap(&originalData)
}
