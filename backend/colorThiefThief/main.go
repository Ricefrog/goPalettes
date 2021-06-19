package main

import (
	"fmt"
	"sort"
	"math"
	"errors"
)

// I am re-implementing the MMCQ (modified median cut quantization) algorithm
// that is implemented in the Python colorthief project on GitHub.
// The colorthief implementation is itself an implementation of an algorithm 
// from the Leptonica library.

const SIGBITS = 5
const RSHIFT = 8 - SIGBITS
const MAX_ITERATION = 1000
const FRACT_BY_POPULATIONS = 0.75

func min(a, b int) int {
	if a <= b {
		return a
	} else {
		return b
	}
	return a
}

func max(a, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

// rgb values transformed into a 15-bit index.
func getColorIndex(r int, g int, b int) int {
	return (r << (2 * SIGBITS)) + (g << SIGBITS) + b
}

// histo is a map that gives the number of pixels in each quantized region
// of color space.
func getHisto(pixels [][]int) map[int]int {
	histo := make(map[int]int)
	for _, pixel := range(pixels) {
		// 8-bit values turned into 5-bit values.
		rval := pixel[0] >> RSHIFT // This is the same as integer division by 8.
		gval := pixel[1] >> RSHIFT
		bval := pixel[2] >> RSHIFT
		index := getColorIndex(rval, gval, bval)
		histo[index] += 1
	}
	return histo
}

func VBoxFromPixels(pixels [][]int, histo map[int]int) *VBox {
	rmin := 1000000
	rmax := 0
	gmin := 1000000
	gmax := 0
	bmin := 1000000
	bmax := 0

	for _, pixel := range(pixels) {
		rval := pixel[0] >> RSHIFT
		gval := pixel[1] >> RSHIFT
		bval := pixel[2] >> RSHIFT
		rmin = min(rval, rmin)
		rmax = max(rval, rmax)
		gmin = min(gval, gmin)
		gmax = max(gval, gmax)
		bmin = min(bval, bmin)
		bmax = max(bval, bmax)
	}

	return &VBox{
		r1: rmin,
		r2: rmax,
		g1: gmin,
		g2: gmax,
		b1: bmin,
		b2: bmax,
		histo: histo,
	}
}

// This function decides how to split each vbox.
func MedianCutApply(histo map[int]int, vbox VBox) (VBox, VBox) {
	// Return nothing if vbox contains no pixels.
	if (vbox.count() == 0) {
		return VBox{invalid: true}, VBox{invalid: true}
	}
	// If only one pixel just return original vbox without splitting.
	if (vbox.count() == 1) {
		return vbox.copy(), VBox{invalid: true}
	}

	rw := vbox.r2 - vbox.r1 + 1
	gw := vbox.g2 - vbox.g1 + 1
	bw := vbox.b2 - vbox.b1 + 1
	maxw := max(rw, max(gw, bw))

	// Find the partial sum arrays along the selected axis.
	total := 0
	sum := 0
	partialSum := make(map[int]int)
	lookAheadSum := make(map[int]int)
	var doCutColor string

	if maxw == rw {
		doCutColor = "r"
		for i := vbox.r1; i <= vbox.r2; i++ {
			sum = 0
			for j := vbox.g1; j <= vbox.g2; j++ {
				for k := vbox.b1; k <= vbox.b2; k++ {
					index := getColorIndex(i, j, k)
					sum += histo[index]
				}
			}
			total += sum
			partialSum[i] = total
		}
	} else if maxw == gw {
		doCutColor = "g"
		for i := vbox.g1; i <= vbox.g2; i++ {
			sum = 0
			for j := vbox.r1; j <= vbox.r2; j++ {
				for k := vbox.b1; k <= vbox.b2; k++ {
					index := getColorIndex(i, j, k)
					sum += histo[index]
				}
			}
			total += sum
			partialSum[i] = total
		}
	} else {
		doCutColor = "b"
		for i := vbox.b1; i <= vbox.b2; i++ {
			sum = 0
			for j := vbox.r1; j <= vbox.r2; j++ {
				for k := vbox.g1; k <= vbox.g2; k++ {
					index := getColorIndex(i, j, k)
					sum += histo[index]
				}
			}
			total += sum
			partialSum[i] = total
		}
	}
	for k, v := range(partialSum) {
		lookAheadSum[k] = total - v
	}

	// Determine the cut planes.
	dim1 := doCutColor + "1"
	dim2 := doCutColor + "2"
	var dim1Val int
	var dim2Val int
	switch doCutColor {
	case "r":
		dim1Val = vbox.r1
		dim2Val = vbox.r2
	case "g":
		dim1Val = vbox.g1
		dim2Val = vbox.g2
	case "b":
		dim1Val = vbox.b1
		dim2Val = vbox.b2
	}
	for i := dim1Val; i <= dim2Val; i++ {
		if partialSum[i] > (total / 2) {
			vbox1 := vbox.copy()
			vbox2 := vbox.copy()
			left := i - dim1Val  // Distance from the lower bound.
			right := dim2Val - i // Distance from the upper bound.
			var d2 int
			if left <= right {
				d2 = min(dim2Val - 1, i + right/2)
			} else {
				d2 = max(dim2Val, i - 1 - left/2)
			}
			// Avoid 0-count boxes.
			for partialSum[d2] == 0 {
				d2 += 1
			}

			count2 := lookAheadSum[d2]
			for count2 == 0 && partialSum[d2-1] == 0 {
				d2 -= 1
				count2 = lookAheadSum[d2]
			}

			// Create function to set attributes based on strings.
			vbox1.setDimWithString(dim2, d2)
			vbox2.setDimWithString(dim1, vbox1.getDimWithString(dim2)+1)
			return vbox1, vbox2
		}
	}
	return VBox{invalid: true}, VBox{invalid: true}
}

// maxColor is the max number of colors to extract.
func Quantize(pixels [][]int, maxColor int) (CMap, error) {
	if len(pixels) == 0 {
		retErr := errors.New("In Quantize: Empty pixel array.\n")
		return CMap{invalid: true}, retErr
	}
	if maxColor < 2 || maxColor > 256 {
		retErr := errors.New("In Quantize: Invalid value for maxColor.")
		return CMap{invalid: true}, retErr
	}

	histo := getHisto(pixels)
	if len(histo) <= maxColor {
		retErr := errors.New("In Quantize: len(histo) <= maxColor.")
		return CMap{invalid: true}, retErr
	}

	// Get the starting vbox from the colors.
	vbox := *VBoxFromPixels(pixels, histo)
	vq := *createVQueue(func(a, b VBox) bool {
		return a.count() < b.count()
	})
	vq.push(vbox)

	// Inner function to do the iteration.
	iter := func(lh VQueue, target float64) error {
		nColor := 1
		nIter := 0
		for nIter < MAX_ITERATION {
			vbox = lh.pop()
			if vbox.count() == 0 {
				lh.push(vbox)
				nIter += 1
				continue
			}
			// Do the cut.
			vbox1, vbox2 := MedianCutApply(histo, vbox)
			if vbox1.invalid {
				return errors.New(
					"In Quantize:iter: Something went wrong when making cut.")
			}
			lh.push(vbox1)
			if !vbox2.invalid {
				lh.push(vbox2)
				nColor += 1
			}
			if float64(nColor) >= target {
				return nil
			}
			if nIter > MAX_ITERATION {
				return nil
			}
			nIter += 1
		}
		return errors.New("In Quantize:iter: No proper return.")
	}

	// First set of colors, sorted by population.
	err := iter(vq, FRACT_BY_POPULATIONS * float64(maxColor))
	if err != nil {
		return CMap{invalid: true}, err
	}

	// Re-sort by the product of pixel occupancy times the size in a color 
	// space.
	vq2 := *createVQueue(func(a, b VBox) bool {
		return (a.count()*a.volume()) < (b.count()*b.volume())
	})
	for vq.size() > 0 {
		vq2.push(vq.pop())
	}

	// Next set: Generate the median cuts using the (npix * vol) sorting.
	err = iter(vq2, float64(maxColor - vq2.size()))
	if err != nil {
		return CMap{invalid: true}, err
	}

	// Calculate the actual colors.
	cmap := *createCMap()
	for vq2.size() > 0 {
		cmap.push(vq2.pop())
	}
	return cmap, nil
}

//------------------------------------------------------------------------------
// 3D colorspace box
type VBox struct {
	r1 int
	r2 int
	g1 int
	g2 int
	b1 int
	b2 int
	histo map[int]int
	invalid bool
}

func (v *VBox) setDimWithString(fstr string, val int) {
	switch fstr {
	case "r1":
		v.r1 = val
	case "r2":
		v.r2 = val
	case "g1":
		v.g1 = val
	case "g2":
		v.g2 = val
	case "b1":
		v.b1 = val
	case "b2":
		v.b2 = val
	}
}

func (v *VBox) getDimWithString(fstr string) int {
	switch fstr {
	case "r1":
		return v.r1
	case "r2":
		return v.r2
	case "g1":
		return v.g1
	case "g2":
		return v.g2
	case "b1":
		return v.b1
	case "b2":
		return v.b2
	default:
		return 42
	}
}

func (v VBox) volume() int {
	sub_r := v.r2 - v.r1
	sub_g := v.g2 - v.g1
	sub_b := v.b2 - v.b1
	return (sub_r + 1) * (sub_g + 1) * (sub_b + 1)
}

func (v VBox) copy() VBox {
	histo := make(map[int]int)
	for k, v := range(v.histo) {
		histo[k] = histo[v]
	}
	return VBox{
		r1: v.r1,
		r2: v.r2,
		g1: v.g1,
		g2: v.g2,
		b1: v.b1,
		b2: v.b2,
		histo: histo,
		invalid: false,
	}
}

func (v VBox) avg() []int {
	ntot := 0
	mult := 1 << (8 - SIGBITS) // 8
	r_sum := 0.0
	g_sum := 0.0
	b_sum := 0.0

	for i := v.r1; i <= v.r2; i++ {
		for j := v.g1; j <= v.g2; j++ {
			for k := v.b1; k <= v.b2; k++ {
				histoIndex := getColorIndex(i, j, k)
				hval := v.histo[histoIndex]
				ntot += hval
				r_sum += float64(hval) * (float64(i) + 0.5) * float64(mult)
				g_sum += float64(hval) * (float64(j) + 0.5) * float64(mult)
				b_sum += float64(hval) * (float64(k) + 0.5) * float64(mult)
			}
		}
	}

	var r_avg int
	var g_avg int
	var b_avg int
	if (ntot > 0) {
		r_avg := int(r_sum / float64(ntot))
		g_avg := int(g_sum / float64(ntot))
		b_avg := int(b_sum / float64(ntot))
	} else {
		r_avg := int(mult * (v.r1 + v.r2 + 1) / 2)
		g_avg := int(mult * (v.g1 + v.g2 + 1) / 2)
		b_avg := int(mult * (v.b1 + v.b2 + 1) / 2)
	}
	return []int{r_avg, g_avg, b_avg}
}

func (v VBox) contains(pixel []int) bool {
	rval := pixel[0] >> RSHIFT
	gval := pixel[1] >> RSHIFT
	bval := pixel[2] >> RSHIFT

	if rval < v.r1 {
		return false
	}
	if rval > v.r2 {
		return false
	}
	if gval < v.g1 {
		return false
	}
	if gval > v.g2 {
		return false
	}
	if bval < v.b1 {
		return false
	}
	if bval > v.b2 {
		return false
	}
	return true
}

func (v VBox) count() int {
	npix := 0
	for i := v.r1; i <= v.r2; i++ {
		for j := v.g1; j <= v.g2; j++ {
			for k := v.b1; k <= v.b2; k++ {
				index := getColorIndex(i, j, k)
				npix += v.histo[index]
			}
		}
	}
	return npix
}
//------------------------------------------------------------------------------
type vbAndColor struct {
	vbox VBox
	color []int
}
// Color map
type CMap struct {
	vBoxes VCQueue
	invalid bool
}

// Initialize the VCQueue passing in the sort key function.
// The function returns the condition for which the first parameter is less
// than the second.
func createCMap() *CMap {
	sortF := func(i, j vbAndColor) bool {
		return (i.vbox.count()*i.vbox.volume())<(j.vbox.count()*j.vbox.volume())
	}
	return &CMap{ vBoxes: *createVCQueue(sortF), invalid: false }
}

// Returns an array of the color arrays of each vbAndColor struct.
func (c CMap) palette() [][]int {
	ret := make([][]int, c.vBoxes.size())
	for i := range(ret) {
		colorArr := c.vBoxes.peek(i).color
		ret[i] = []int{colorArr[0], colorArr[1], colorArr[2]}
	}
	return ret
}

func (c *CMap) push(vbox VBox) {
	newVbc := vbAndColor{
		vbox: vbox,
		color: vbox.avg(),
	}
	c.vBoxes.push(newVbc)
}

func (c CMap) size() int {
	return c.vBoxes.size()
}

func (c CMap) nearest(color []int) []int {
	var d1 float64 = 42.0000000042
	p_color := make([]int, 3)

	for i := 0; i < c.vBoxes.size(); i++ {
		vbox := c.vBoxes.peek(i)
		d2 := math.Sqrt(
			math.Pow(float64(color[0] - vbox.color[0]), 2) +
			math.Pow(float64(color[1] - vbox.color[1]), 2) +
			math.Pow(float64(color[2] - vbox.color[2]), 2),
		)
		if d1 == 42.0000000042 || d2 < d1 {
			tempC := vbox.color
			d1 = d2
			p_color = []int{tempC[0], tempC[1], tempC[2]}
		}
	}
	return p_color
}

/*
// Returns the average color of the vBox that contains the color parameter.
// If none of the existing vBoxes contain the color, return the average color 
// of the nearest vBox.
func (c CMap) map(color []int) []int {
	for i := range(c.vBoxes.size()) {
		vbox := c.vBoxes.peek(i)
		if vbox.contains(color):
			return vbox.color
	}
	return c.nearest(color)
}
*/

//------------------------------------------------------------------------------
// Priority queue for vBoxes
type vq_sort_key func(VBox, VBox) bool
type vq_mapFunction func(VBox) VBox

type VQueue struct {
	sortKey vq_sort_key
	contents []VBox
	sorted bool
}

func createVQueue(key vq_sort_key) *VQueue {
	return &VQueue{
		sortKey: key,
		contents: make([]VBox, 0),
		sorted: false,
	}
}

func (vq *VQueue) sort() {
	sort.Slice(vq.contents, vq.sortKey)
	vq.sorted = true
}

func (vq *VQueue) push(el VBox) {
	vq.contents := append(vq.contents, el)
	vq.sorted = false
}

func (vq *VQueue) peek(index int) VBox {
	if !vq.sorted {
		vq.sort()
	}
	return vq.contents[index]
}

func (vq *VQueue) pop() VBox {
	if !vq.sorted {
		vq.sort()
	}
	ret := vq.contents[len(vq.contents) - 1]
	vq.contents = vq.contents[:len(vq.contents)]
	return ret
}

func (vq VQueue) size() int {
	return len(vq.contents)
}

func (vq VQueue) mapFunc(f_to_use vq_mapFunction) {
	retArr := []VBox
	for i, el := range(vq.contents) {
		retArr[i] :=  f_to_use(el)
	}
}

//------------------------------------------------------------------------------
// Priority queue for vbAndColor structs
type vcq_sort_key func(vbAndColor, vbAndColor) bool
type vcq_mapFunction func(vbAndColor) vbAndColor

type VCQueue struct {
	sortKey vcq_sort_key
	contents []vbAndColor
	sorted bool
}

func createVCQueue(key vcq_sort_key) *VCQueue {
	return &VCQueue{
		sortKey: key,
		contents: make([]vbAndColor),
		sorted: false,
	}
}

func (vcq *VCQueue) sort() {
	sort.Slice(vcq.contents, vcq.sortKey)
	vcq.sorted = true
}

func (vcq *VCQueue) push(el vbAndColor) {
	vcq.contents := append(vcq.contents, el)
	vcq.sorted = false
}

func (vcq *VCQueue) peek(index int) vbAndColor {
	if !vcq.sorted {
		vcq.sort()
	}
	return vcq.contents[index]
}

func (vcq *VCQueue) pop() vbAndColor {
	if !vcq.sorted {
		vcq.sort()
	}
	ret := vcq.contents[len(vcq.contents) - 1]
	vcq.contents = vcq.contents[:len(vcq.contents)]
	return ret
}

func (vcq VCQueue) size() int {
	return len(vcq.contents)
}

func (vcq VCQueue) mapFunc(f_to_use vcq_mapFunction) {
	retArr := []vbAndColor
	for i, el := range(vcq.contents) {
		retArr[i] :=  f_to_use(el)
	}
}

func main() {
	fmt.Println("Works!")
}
