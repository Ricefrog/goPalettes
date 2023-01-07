package main

import (
	"goPalettes/imageManip"
	"os"
	"log"
	"image"
	"fmt"
	"strings"
	"strconv"
)

var test image.Image

func main() {
	//imageManip.Stub_1()
	//diffs()
	//testColStringToArr()
	//fmt.Println(test)
	//fmt.Println(test == nil)
	//testingRgbaToHex()
	//splittingColFreqMapTesting()
	/*
	testMap := make(map[string]int)
	mapTest(testMap)
	fmt.Println(testMap)
	*/
	testArrMap := make([]map[string]int, 5)
	for i := 0; i < 5; i++ {
		testArrMap[i] = make(map[string]int)
	}
	mapArrTest(testArrMap)
	fmt.Println(testArrMap)
}

func mapArrTest(m []map[string]int) {
	for i, m := range m {
		m["testing"] = i
	}
}

func mapTest(m map[string]int) {
	foo := "testing"
	m[foo] = 15
}

/*
func splittingColFreqMapTesting() {
	testColFreqMap := make(map[string]int)
	sections := 4

	for i := 0; i < 10; i++ {
		testColFreqMap[string(i+65)] = i
	}

	retArr := imageManip.SplitColFreqMap(sections, testColFreqMap)
	for _, m := range retArr {
		fmt.Printf("\n%v\n", m)
	}
}
*/

func rgbaToHex(rgba string) string {
	temp := strings.Split(rgba, " ")[:3]
	fmt.Println("temp:", temp)
	lt := []int{len(temp[0]), len(temp[1]), len(temp[2])}

	r, g, b := temp[0][1:lt[0]-1], temp[1][:lt[1]-1], temp[2][:lt[2]-1]
	fmt.Printf("r: %s, g: %s, b: %s\\", r, g, b)

	ri, _ := strconv.Atoi(r)
	gi, _ := strconv.Atoi(g)
	bi, _ := strconv.Atoi(b)
	fmt.Println("After Atoi:")
	fmt.Printf("ri: %d, gi: %d, bi: %d\\", ri, gi, bi)

	rh := strconv.FormatInt(int64(ri), 16)
	if len(rh) == 1 {
		rh = "0"+rh
	}
	gh := strconv.FormatInt(int64(gi), 16)
	if len(gh) == 1 {
		gh = "0"+gh
	}
	bh := strconv.FormatInt(int64(bi), 16)
	if len(bh) == 1 {
		bh = "0"+bh
	}

	ret := "#"+rh+gh+bh
	return ret
}

func testingRgbaToHex() {
	testString := "(198, 255, 255, 255)"
	fmt.Println(rgbaToHex(testString))
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
