package main

import (
	"fmt"
	"sort"
	"sqrt"
)

const SIGBITS = 5
const RSHIFT = 8 - SIGBITS
const MAX_ITERATION = 1000
const FRACT_BY_POPULATIONS = 0.75

// I am re-implementing the MMCQ (modified median cut quantization) algorithm
// that is implemented in the Python colorthief project on GitHub.
// The colorthief implementation is itself an implementation of an algorithm 
// from the Leptonica library.

// rgb values transformed into a 15-bit index.
func getColorIndex(r int, g int, b int) int {
	return (r << (2 * SIGBITS)) + (g << SIGBITS) + b
}

// histo is a map that gives the number of pixels in each quantized region
// of color space.
func getHisto(pixels [][]int) map[int]int {
	histo := make(map[int]int)
	for pixel := range(pixels) {
		// 8-bit values turned into 5-bit values.
		rval := pixel[0] >> RSHIFT // This is the same as integer division by 8.
		gval := pixel[1] >> RSHIFT
		bval := pixel[2] >> RSHIFT
		index := getColorIndex(rval, gval, bval)
		histo[index] += 1
	}
	return histo
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
		histo,
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
				r_sum += hval * (i + 0.5) * mult
				g_sum += hval * (j + 0.5) * mult
				b_sum += hval * (k + 0.5) * mult
			}
		}
	}

	if (ntot > 0) {
		r_avg := int(r_sum / ntot)
		g_avg := int(g_sum / ntot)
		b_avg := int(b_sum / ntot)
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
		return False
	}
	if rval > v.r2 {
		return False
	}
	if gval < v.g1 {
		return False
	}
	if gval > v.g2 {
		return False
	}
	if bval < v.b1 {
		return False
	}
	if bval > v.b2 {
		return False
	}
	return True
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
}

// Initialize the VCQueue passing in the sort key function.
// The function returns the condition for which the first parameter is less
// than the second.
func createCMap() *CMap {
	return &CMap{
		vBoxes: createVCQueue(func (i, j vbAndColor) bool {
			return (i.vbox.count*i.vbox.volume) < (j.vbox.count*j.vbox.volume)
		})
	}
}

// Returns an array of the color arrays of each vbAndColor struct.
func (c CMap) palette() [][]int {
	ret := make([][]int, len(c.vBoxes))
	for i := range(ret) {
		colorArr := c.vBoxes[i].color
		ret[i] := int{colorArr...}
	}
}

func (c *CMap) push(vbox VBox) {
	newVbc := vbAndColor{
		vbox,
		color: vbox.avg(),
	}
	c.vBoxes = append(c.vBoxes, newVbc)
}

func (c CMap) size() int {
	return c.vBoxes.size()
}

func (c CMap) nearest(color []int) []int {
	var d1 float64 = nil
	p_color := make([]int, 3)

	for i := range(c.vBoxes.size()) {
		vbox := c.vBoxes.peek(i)
		d2 := math.Sqrt(
			math.Pow(color[0] - vbox.color[0], 2) +
			math.Pow(color[1] - vbox.color[1], 2) +
			math.Pow(color[2] - vbox.color[2], 2)
		)
		if d1 == nil || d2 < d1 {
			d1 = d2
			p_color = int{vbox.color...}
		}
	}
	return p_color
}

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

//------------------------------------------------------------------------------
// Priority queue for vBoxes
type vq_sort_key func(int int) bool
type vq_mapFunction func(VBox) VBox

type VQueue struct {
	sortKey vq_sort_key
	contents []VBox
	sorted bool
}

func createVQueue(key vq_sort_key) *VQueue {
	return &VQueue{
		sortKey: key,
		contents: make([]VBox),
		sorted: False,
	}
}

func (vq *VQueue) sort() {
	sort.Slice(vq.contents, sortKey)
	vq.sorted = True
}

func (vq *VQueue) push(el VBox) {
	vq.contents := append(vq.contents, el)
	vq.sorted = False
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
		sorted: False,
	}
}

func (vcq *VCQueue) sort() {
	sort.Slice(vcq.contents, vcq.sortKey)
	vcq.sorted = True
}

func (vcq *VCQueue) push(el VBox) {
	vcq.contents := append(vcq.contents, el)
	vcq.sorted = False
}

func (vcq *VCQueue) peek(index int) VBox {
	if !vcq.sorted {
		vcq.sort()
	}
	return vcq.contents[index]
}

func (vcq *VCQueue) pop() VBox {
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
	retArr := []VBox
	for i, el := range(vcq.contents) {
		retArr[i] :=  f_to_use(el)
	}
}
