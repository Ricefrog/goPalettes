package imageManip

import (
	"fmt"
	"image"
	_ "image/png"
	_ "image/jpeg"
	"image/color"
	"os"
	"errors"
	"log"
	"sync"
)

var IMAGE_PATH = "./images/"
var CORES_TO_USE = 4

type Domain struct {
	xUpper int
	xLower int
	yUpper int
	yLower int
}

type ColAndFreq struct {
	colString string
	frequency int
}

func ParseArgs() (s string, e string, err error) {
	args := os.Args[1:]
	if len(args) < 2 {
		err = errors.New("Not enough arguments.")
		return
	}

	s = args[0]
	e = args[1]
	return
}

func CreateColorFrequencyMap(img image.Image) (map[string]int) {
	colFreqMap := make(map[string]int)
	domains := CreateDomains(img)
	allMaps := make([]map[string]int, CORES_TO_USE)
	fmt.Println(domains)

	var wg sync.WaitGroup
	wg.Add(CORES_TO_USE)

	for i, dom := range domains {
		newMap := make(map[string]int)
		go CountColors(img, dom, newMap, &wg)
		allMaps[i] = newMap
	}
	wg.Wait()

	MergeColorFrequencyMaps(colFreqMap, allMaps)
	return colFreqMap
}

// Split the bounds of the image into equal parts based on the
// number of cores being utilized.
func CreateDomains(img image.Image) (domains []Domain) {
	xChunk := img.Bounds().Max.X / CORES_TO_USE
	maxY := img.Bounds().Max.Y

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
		xUpper: img.Bounds().Max.X,
		yUpper: maxY,
		xLower: (xChunk*(CORES_TO_USE-1)),
		yLower: 0,
	}
	domains = append(domains, lastDom)

	return
}

// Add to the map for 
func CountColors(img image.Image, dom Domain,
newMap map[string]int, wg *sync.WaitGroup) (map[string]int) {
	defer wg.Done()

	for i := dom.xLower; i < dom.xUpper; i++ {
		for j := dom.yLower; j < dom.yUpper; j++ {
			colString := ColorToString(img.At(i, j))
			newMap[colString]++
		}
	}
	return newMap
}

func ColorToString(col color.Color) string {
	colAsserted, _ := col.(color.NRGBA) // Type assertion
	retString := fmt.Sprintf("(%d, %d, %d, %d)",
	colAsserted.R, colAsserted.G, colAsserted.B, colAsserted.A)
	return retString
}

func MergeColorFrequencyMaps(masterMap map[string]int, maps []map[string]int) {
	for _, curMap := range maps {
		for key, el := range curMap {
			masterMap[key] += el
		}
	}
}

func MostProminentColor(colFreqMap map[string]int) ColAndFreq {
	max := ColAndFreq{
		colString: "placeholder",
		frequency: 0,
	}

	for key, el := range colFreqMap {
		if el > max.frequency {
			max.colString = key
			max.frequency = el
		}
	}

	return max
}

// return an array of the n most prominent colors
func GetMostProminentColors(n int, colFreqMap map[string]int) []ColAndFreq {
	ret := make([]ColAndFreq, n)
	for i := 0; i < n; i++ {
		cur := MostProminentColor(colFreqMap)
		ret[i] = cur
		delete(colFreqMap, cur.colString)
	}
	return ret
}

func Stub_1() {
	s, e, err := ParseArgs()
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
	colorFrequencyMap := CreateColorFrequencyMap(originalData)
	fmt.Println(colorFrequencyMap)
	fmt.Printf("%d unique colors found.\n", len(colorFrequencyMap))
	fmt.Printf("Most prominent color: %v\n",
	MostProminentColor(colorFrequencyMap))

	n := 5
	fmt.Printf("%d most prominent colors: %v", n,
	GetMostProminentColors(n, colorFrequencyMap))
}
