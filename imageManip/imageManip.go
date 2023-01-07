package imageManip

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"strconv"
	"strings"
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
	ColString string `json:"color"`
	Frequency int    `json:"frequency"`
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

func CreateColorFrequencyMap(img image.Image) map[string]int {
	colFreqMap := make(map[string]int)
	domains := CreateDomains(img)
	allMaps := make([]map[string]int, CORES_TO_USE)
	fmt.Println("domains:", domains)

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
			xUpper: xChunk * i,
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
		xLower: (xChunk * (CORES_TO_USE - 1)),
		yLower: 0,
	}
	domains = append(domains, lastDom)

	return
}

// Add to the map for
func CountColors(img image.Image, dom Domain,
	newMap map[string]int, wg *sync.WaitGroup) map[string]int {
	defer wg.Done()

	for i := dom.xLower; i < dom.xUpper; i++ {
		for j := dom.yLower; j < dom.yUpper; j++ {
			ColString := ColorToString(img.At(i, j))
			newMap[ColString]++
		}
	}
	return newMap
}

func ColorToString(col color.Color) string {
	r, g, b, a := col.RGBA()
	retString := fmt.Sprintf("%d, %d, %d, %d",
		uint8(r), uint8(g), uint8(b), uint8(a))
	return retString
}

func MergeColorFrequencyMaps(masterMap map[string]int, maps []map[string]int) {
	for _, curMap := range maps {
		for key, el := range curMap {
			masterMap[key] += el
		}
	}
}

func mostProminentColor(colFreqMap map[string]int) ColAndFreq {
	max := ColAndFreq{
		ColString: "placeholder",
		Frequency: 0,
	}

	for key, el := range colFreqMap {
		if el > max.Frequency {
			max.ColString = key
			max.Frequency = el
		}
	}

	return max
}

// return an array of the n most prominent colors
func GetMostProminentColors(n int, colFreqMap map[string]int) []ColAndFreq {
	ret := make([]ColAndFreq, n)
	for i := 0; i < n; i++ {
		cur := mostProminentColor(colFreqMap)
		ret[i] = cur
		delete(colFreqMap, cur.ColString)
	}
	return ret
}

// Takes an array of ColAndFreq and creates a composite color using
// a weighted average. Returns a ColAndFreq struct with this composite
// color and the sum of all frequencies in the array.
func getCompositeColor(colors []ColAndFreq) ColAndFreq {
	Frequency := 0
	r, g, b := 0.0, 0.0, 0.0
	for _, col := range colors {
		// for each element convert string into array and add to
		// color sums.
		colArr := ColStringToArr(col.ColString)
		r += colArr[0] * float64(col.Frequency)
		g += colArr[1] * float64(col.Frequency)
		b += colArr[2] * float64(col.Frequency)
		Frequency += col.Frequency
	}

	r /= float64(Frequency)
	g /= float64(Frequency)
	b /= float64(Frequency)
	rS, gS, bS := strconv.Itoa(int(r)), strconv.Itoa(int(g)), strconv.Itoa(int(b))
	tempStrArr := []string{rS, gS, bS, "255"}
	ColString := strings.Join(tempStrArr, ", ")

	composite := ColAndFreq{
		ColString,
		Frequency,
	}

	return composite
}

// Gets the most prominent color.
// Makes an array of all the colors in the map that are within the
// tolerance value. Deletes these similar colors from colFreqMap.
// Combines the colors in the array into one. Uses weighted average.
// Returns this new composite color.
// At the end of the function colFreqMap has lost its most prominent color
// and all colors similar to it.
func mostProminentColorImproved(colFreqMap map[string]int, tolerance float64) ColAndFreq {
	mostProminent := mostProminentColor(colFreqMap)
	mPCol := mostProminent.ColString
	delete(colFreqMap, mostProminent.ColString)

	similarColors := []ColAndFreq{mostProminent}

	for k, v := range colFreqMap {
		if distance(ColStringToArr(mPCol), ColStringToArr(k)) < tolerance {
			similarColor := ColAndFreq{
				ColString: k,
				Frequency: v,
			}
			similarColors = append(similarColors, similarColor)
			delete(colFreqMap, k)
		}
	}
	return getCompositeColor(similarColors)
}

// Takes a colFreqMap and applies a tolerance value to get the specified
// number of 'most prominent colors'. Each of these prominent colors is
// a weighted average of all the colors similar to it.
func getMostProminentColorsImproved(
	numberOfColors int,
	colFreqMap map[string]int,
	tolerance float64,
) []ColAndFreq {
	ret := make([]ColAndFreq, numberOfColors)
	for i := 0; i < numberOfColors; i++ {
		cur := mostProminentColorImproved(colFreqMap, tolerance)
		ret[i] = cur
	}
	return ret
}

// convert rgb values in string to an array of three
func ColStringToArr(str string) (retArr [3]float64) {
	arr := strings.Split(str, " ")
	//fmt.Println(arr)
	arr = arr[:3]
	for i, val := range arr {
		temp, _ := strconv.Atoi(val[:len(val)-1])
		retArr[i] = float64(temp)
	}
	return
}

// create groups of similar colors according to some distance tolerance value
func SimplifyColFreqMap(
	tolerance float64,
	colFreqMap map[string]int,
) map[string]int {
	fmt.Println("SimplifyColMap was called.")
	// the keys of the map act as representatives of the color group
	colorGroups := make(map[string][]ColAndFreq)

	for k, v := range colFreqMap {
		//fmt.Println("Outer loop.")
		groupFound := false
		newMember := ColAndFreq{
			ColString: k,
			Frequency: v,
		}
		for rep, _ := range colorGroups {
			//fmt.Println("Inner loop.")
			// if a color fits into a color group add it to the array
			// and delete it from the original map.
			if distance(ColStringToArr(rep), ColStringToArr(k)) < tolerance {
				//fmt.Println("inner if entered.")
				colorGroups[rep] = append(colorGroups[rep], newMember)
				//delete(colFreqMap, k)
				groupFound = true
				break
			}
		}
		// if the color couldn't find a group to fit into, create
		// a new color group with that color as the rep
		if !groupFound {
			//fmt.Println("outer if entered.")
			colorGroups[k] = []ColAndFreq{newMember}
		}
	}
	fmt.Printf("Loops exited. %d color groups created.\n", len(colorGroups))
	// merge color groups into a return color Frequency map.
	return mergeColorGroups(colorGroups)
}

// This removes all elements in colFreqMap that are below the
// frequency threshold.
func flenseColFreqMap(colFreqMap map[string]int) {
	threshold := 100
	for key, val := range colFreqMap {
		if val < threshold {
			delete(colFreqMap, key)
		}
	}
}

// create groups of similar colors according to some distance tolerance value
func SimplifyColFreqMapConcurrent(
	tolerance float64,
	colFreqMap map[string]int,
	numberOfGoroutines int,
) map[string]int {
	fmt.Println("SimplifyColMapConcurrent was called.")

	fmt.Printf("\nLength of colFreqMap before flensing: %d\n", len(colFreqMap))
	flenseColFreqMap(colFreqMap)
	fmt.Printf("\nLength of colFreqMap after flensing: %d\n", len(colFreqMap))

	// Split map into sections to be handled concurrently.
	// Each subMap maps a color value to its frequency in the image.
	numberOfSections := numberOfGoroutines
	subMaps := splitColFreqMap(numberOfSections, colFreqMap)

	// Comparing the size of the submaps to the main colFreqMap.
	for i, s := range subMaps {
		fmt.Printf("Length of subMap %d: %d\n", i, len(s))
	}

	debugSubMaps(subMaps)

	// each element in colorGroupsArray holds the corresponding colorGroups
	// for each subMap.
	colorGroupsArray := make([]map[string][]ColAndFreq, numberOfSections)
	for i := 0; i < numberOfSections; i++ {
		colorGroupsArray[i] = make(map[string][]ColAndFreq)
	}

	// getColorGroups is the performance bottleneck.
	var wg sync.WaitGroup
	wg.Add(numberOfSections)
	for i, subMap := range subMaps {
		go getColorGroups(tolerance, subMap, colorGroupsArray, i, &wg)
	}
	wg.Wait()

	// Now merge colorGroupsArray into a single color group.
	// Maybe: Use tolerance value to let similar groups swallow each other.
	colorGroups := make(map[string][]ColAndFreq)
	for _, subGroup := range colorGroupsArray {
		for k, v := range subGroup {
			colorGroups[k] = v
		}
	}

	fmt.Printf(
		"subMaps merged. %d color groups created.\n",
		len(colorGroups),
	)

	// merge color groups into a return color Frequency map. (map[string]int).
	retMap := mergeColorGroups(colorGroups)

	fmt.Printf(
		"Color groups merged. %d colors in return map.\n",
		len(retMap),
	)

	return retMap
}

// Split colFreqMap into an array of submaps. Flensed.
func splitColFreqMap(sections int, colFreqMap map[string]int) []map[string]int {
	ret := make([]map[string]int, sections)
	currentSection := 0
	counter := 1
	subSectionLength := len(colFreqMap) / sections

	// intialize maps
	for i := 0; i < sections; i++ {
		ret[i] = make(map[string]int)
	}

	for k, v := range colFreqMap {
		if currentSection < sections-1 && counter > subSectionLength {
			currentSection++
			counter = 1
		}
		ret[currentSection][k] = v
		counter++
	}
	return ret
}

func getColorGroups(
	tolerance float64,
	colFreqMap map[string]int,
	colorGroupsArray []map[string][]ColAndFreq,
	index int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	// the keys of the colorGroups map act as representatives of the color group
	// The values of the keys of colorGroups are arrays of ColAndFreq structs.
	colorGroups := colorGroupsArray[index]

	for k, v := range colFreqMap {
		//fmt.Println("Outer loop.")
		groupFound := false
		newMember := ColAndFreq{
			ColString: k,
			Frequency: v,
		}
		// Observation: The higher the tolerance, the faster the program runs.
		// Why is this? I do not know.
		for rep, _ := range colorGroups {
			// if a color fits into a color group add it to the array.
			if distance(ColStringToArr(rep), ColStringToArr(k)) < tolerance {
				colorGroups[rep] = append(colorGroups[rep], newMember)
				groupFound = true
				break
			}
		}
		// if the color couldn't find a group to fit into, create
		// a new color group with that color as the rep
		if !groupFound {
			//fmt.Println("outer if entered.")
			colorGroups[k] = []ColAndFreq{newMember}
		}
	}

	fmt.Printf(
		"%d color groups created for subMap %d.\n",
		len(colorGroups),
		index,
	)
}

// used this for debugging
func countEmptyStrings(colorGroups map[string][]ColAndFreq) {
	count := 0
	emptyArrs := make(map[string][]ColAndFreq)
	for k, v := range colorGroups {
		for _, col := range v {
			if col.ColString == "" {
				count++
				fmt.Println(col)
				emptyArrs[k] = v
			}
		}
	}
	fmt.Printf("%d empty elements found.\n", count)
	fmt.Println(emptyArrs)
}

func mergeColorGroups(
	colorGroups map[string][]ColAndFreq,
) map[string]int {
	fmt.Println("mergeColorGroups called.")
	merged := make(map[string]int)
	for _, v := range colorGroups {
		retVal := mergeColAndFreqArr(v)
		merged[retVal.ColString] = retVal.Frequency
	}
	return merged
}

// takes an array of ColAndFreq, gets its average color and sums the
// frequencies of each element.
func mergeColAndFreqArr(cols []ColAndFreq) ColAndFreq {
	totalColors := float64(len(cols))
	Frequency := 0
	r, g, b := 0.0, 0.0, 0.0
	for _, col := range cols {
		// for each element convert string into array and add to
		// color sums.
		//fmt.Printf("\"%s\"\n", col.ColString)
		if col.ColString == "" {
			fmt.Println(col)
		}
		//
		colArr := ColStringToArr(col.ColString)
		r += colArr[0]
		g += colArr[1]
		b += colArr[2]
		Frequency += col.Frequency
	}

	r /= totalColors
	g /= totalColors
	b /= totalColors
	rS, gS, bS := strconv.Itoa(int(r)), strconv.Itoa(int(g)), strconv.Itoa(int(b))
	tempStrArr := []string{rS, gS, bS, "255"}
	ColString := strings.Join(tempStrArr, ", ")

	retColAndFreq := ColAndFreq{
		ColString,
		Frequency,
	}

	return retColAndFreq
}

func distance(p1 [3]float64, p2 [3]float64) float64 {
	return math.Sqrt(sq(p1[0]-p2[0]) + sq(p1[1]-p2[1]) + sq(p1[2]-p2[2]))
}

func sq(a float64) float64 {
	return a * a
}

func rgbaToHex(rgba string) string {
	temp := strings.Split(rgba, " ")[:3]
	lt := []int{len(temp[0]), len(temp[1]), len(temp[2])}

	r, g, b := temp[0][1:lt[0]-1], temp[1][:lt[1]-1], temp[2][:lt[2]-1]
	ri, _ := strconv.Atoi(r)
	gi, _ := strconv.Atoi(g)
	bi, _ := strconv.Atoi(b)

	rh := strconv.FormatInt(int64(ri), 16)
	if len(rh) == 1 {
		rh = "0" + rh
	}
	gh := strconv.FormatInt(int64(gi), 16)
	if len(gh) == 1 {
		gh = "0" + gh
	}
	bh := strconv.FormatInt(int64(bi), 16)
	if len(bh) == 1 {
		bh = "0" + bh
	}

	ret := "#" + rh + gh + bh
	return ret
}

// convert an array of ColAndFreq structs to have hex codes for strings.
func rgbaToHexArr(arr []ColAndFreq) []ColAndFreq {
	for i := range arr {
		arr[i].ColString = rgbaToHex(arr[i].ColString)
	}
	return arr
}

func removeContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.Remove(name)
		if err != nil {
			return err
		}
	}

	return nil
}

// This function is being used to help determine if each subMap has unique
// elements.
func debugSubMaps(subMaps []map[string]int) {
	err := removeContents("subMaps")
	if err != nil {
		fmt.Printf("\nSomething went wrong removing contents of subMaps\n")
		fmt.Println(err)
		return
	}

	for i, subMap := range subMaps {
		filePath := fmt.Sprintf("subMaps/subMap%d", i)
		f, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("\nError creating %s\n", filePath)
			fmt.Println(err)
			return
		}

		data := []byte("")
		for key, value := range subMap {
			temp := fmt.Sprintf("%s: %d\n", key, value)
			data = append(data, []byte(temp)...)
		}

		_, err = f.Write(data)
		if err != nil {
			fmt.Printf("\nError writing to %s\n", filePath)
			fmt.Println(err)
			return
		}

		f.Close()
	}
}

func ExtractPalette(
	uploaded image.Image,
	colsToExtract int,
	tolerance float64,
) []ColAndFreq {
	colorFrequencyMap := CreateColorFrequencyMap(uploaded)
	colorFrequencyMap = SimplifyColFreqMap(tolerance, colorFrequencyMap)
	return rgbaToHexArr(GetMostProminentColors(colsToExtract, colorFrequencyMap))
}

func ExtractPaletteConcurrent(
	uploaded image.Image,
	colsToExtract int,
	numberOfGoroutines int,
	tolerance float64,
) []ColAndFreq {
	colorFrequencyMap := CreateColorFrequencyMap(uploaded)
	colorFrequencyMap = SimplifyColFreqMapConcurrent(
		tolerance,
		colorFrequencyMap,
		numberOfGoroutines,
	)
	return rgbaToHexArr(getMostProminentColorsImproved(colsToExtract, colorFrequencyMap, tolerance))
}
