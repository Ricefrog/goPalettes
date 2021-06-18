package main

import (
	"fmt"
	"sort"
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

func (v VBox) avg() (int, int, int) {
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
	return r_avg, g_avg, b_avg
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
// Color map
type CMap struct {
	vBoxes PQueue
}

//------------------------------------------------------------------------------
// Priority queue for vBoxes
type sort_key func(int int) bool
type mapFunction func(VBox) VBox

type VQueue struct {
	sortKey sort_key
	contents []VBox
	sorted bool
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

func (vq VQueue) mapFunc(f_to_use mapFunction) {
	retArr := []VBox
	for i, el := range(vq.contents) {
		retArr[i] :=  f_to_use(el)
	}
}

// The median cut algorithm needs an array of rgb pixel slices.

