package imageManip

import (
	"fmt"
	"image"
	"os"
	"testing"
)

var filePath string = "../images/botnsCover.jpg"

func BenchmarkMedianCut(b *testing.B) {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file (%s): %s\n", filePath, err.Error())
		return
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Printf("Error decoding image (%s): %s\n", filePath, err.Error())
		return
	}

	for i := 0; i < b.N; i++ {
		GetPaletteMC(&img, 4)
	}
}

// Goroutines for initial split.
func BenchmarkMedianCutConcurrent1(b *testing.B) {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file (%s): %s\n", filePath, err.Error())
		return
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Printf("Error decoding image (%s): %s\n", filePath, err.Error())
		return
	}

	for i := 0; i < b.N; i++ {
		GetPaletteMC1(&img, 4)
	}
}

// Goroutines for every split.
func BenchmarkMedianCutConcurrent2(b *testing.B) {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file (%s): %s\n", filePath, err.Error())
		return
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Printf("Error decoding image (%s): %s\n", filePath, err.Error())
		return
	}

	for i := 0; i < b.N; i++ {
		GetPaletteMC2(&img, 4)
	}
}
