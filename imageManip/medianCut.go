package imageManip

import (
	"fmt"
	"image"
	"math"
	"sort"
	"sync"
)

type pix struct {
	r uint64
	g uint64
	b uint64
}

func (p pix) HexString() string {
	return fmt.Sprintf("#%02x%02x%02x", p.r, p.g, p.b)
}

func getRanges(img_arr []pix) (uint64, uint64, uint64) {
	var (
		rmin uint64 = ^uint64(0)
		rmax uint64 = 0
		gmin uint64 = ^uint64(0)
		gmax uint64 = 0
		bmin uint64 = ^uint64(0)
		bmax uint64 = 0
	)

	for _, p := range img_arr {
		if p.r < rmin {
			rmin = p.r
		}

		if p.r > rmax {
			rmax = p.r
		}
	}

	for _, p := range img_arr {
		if p.g < gmin {
			gmin = p.g
		}

		if p.g > gmax {
			gmax = p.g
		}
	}

	for _, p := range img_arr {
		if p.b < bmin {
			bmin = p.b
		}

		if p.b > bmax {
			bmax = p.b
		}
	}

	return rmax - rmin, gmax - gmin, bmax - bmin
}

func average(arr []pix) pix {
	var (
		rsum uint64
		gsum uint64
		bsum uint64

		ravg uint64
		gavg uint64
		bavg uint64
	)
	l := uint64(len(arr))

	for _, p := range arr {
		rsum += p.r
		gsum += p.g
		bsum += p.b
	}

	ravg = rsum / l
	gavg = gsum / l
	bavg = bsum / l

	return pix{
		r: ravg,
		g: gavg,
		b: bavg,
	}
}

func flatten(img image.Image) []pix {
	mX, mY := img.Bounds().Max.X, img.Bounds().Max.Y
	arraySize := mX * mY

	flat := make([]pix, arraySize)
	for j := 0; j < mY; j++ {
		for i := 0; i < mX; i++ {
			col := img.At(i, j)
			r, g, b, _ := col.RGBA()

			// 16-bit to 8-bit
			r = r >> 8
			g = g >> 8
			b = b >> 8

			flat[mX*j+i] = pix{uint64(r), uint64(g), uint64(b)}
		}
	}

	return flat
}

func split(img_arr []pix, depth uint, palette []string, index *uint, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}

	if len(img_arr) == 0 {
		fmt.Printf("Length of img_arr is 0. Returning.\n")
		return
	}

	if depth == 0 {
		hexCode := average(img_arr).HexString()
		palette[*index] = hexCode
		*index++
		return
	}

	var highestRange int
	rRange, gRange, bRange := getRanges(img_arr)
	if rRange >= gRange && rRange >= bRange {
		highestRange = 0
	} else if gRange >= rRange && gRange >= bRange {
		highestRange = 1
	} else {
		highestRange = 2
	}

	switch highestRange {
	case 0:
		sort.SliceStable(img_arr, func(i, j int) bool { return img_arr[i].r < img_arr[j].r })
	case 1:
		sort.SliceStable(img_arr, func(i, j int) bool { return img_arr[i].g < img_arr[j].g })
	case 2:
		sort.SliceStable(img_arr, func(i, j int) bool { return img_arr[i].b < img_arr[j].b })
	}

	medianIndex := len(img_arr) / 2

	var wgInner sync.WaitGroup
	wgInner.Add(2)
	go split(img_arr[:medianIndex], depth-1, palette, index, &wgInner)
	go split(img_arr[medianIndex:], depth-1, palette, index, &wgInner)
	wgInner.Wait()
}

// returns hexcodes for palette of num colors where num is a power of 2
func GetPaletteMC(img *image.Image, n uint) []string {
	flat := flatten(*img)

	var index uint
	palette := make([]string, int(math.Pow(2, float64(n))))

	split(flat, n, palette, &index, nil)
	return palette
}
