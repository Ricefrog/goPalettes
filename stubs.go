package main

import (
	"goPalettes/imageManip"
	"os"
	"log"
	"image"
	"fmt"
)

func main() {
	imageManip.Stub_1()
	//diffs()
	//testColStringToArr()
}

func testColStringToArr() {
	fmt.Println(imageManip.ColStringToArr("23, 233, 90, 500"))
}

// test differences between png and jpeg
func diffs() {
	IMAGE_PATH := imageManip.IMAGE_PATH
	s, e, err := imageManip.ParseArgs()
	// s is jpeg and e is png
	jpgFile, err := os.Open(IMAGE_PATH+s)
	if err != nil {
		log.Fatal(err)
	}
	defer jpgFile.Close()
	jpgData, _, err := image.Decode(jpgFile)

	pngFile, err := os.Open(IMAGE_PATH+e)
	if err != nil {
		log.Fatal(err)
	}
	defer pngFile.Close()
	pngData, _, err := image.Decode(pngFile)

	fmt.Printf("Type of jpgData: %T\n", jpgData)
	fmt.Printf("Type of pngData: %T\n", pngData)
	fmt.Println(jpgData.At(0, 0))
	fmt.Println(pngData.At(0, 0))
}
