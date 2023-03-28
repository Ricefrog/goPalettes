package imageManip

import (
	"errors"
	"fmt"
	"image"
	"sort"
)

type pix struct {
	r uint32
	g uint32
	b uint32
}

func notPowerOfTwo(n uint) bool {
	i := 0
	for {
		power := uint(1 << i)
		if n == power {
			return true
		}
		if n < power {
			break
		}
	}
	return false
}

func getRanges(img_arr []pix) (uint32, uint32, uint32) {
	var (
		rmin uint32 = ^uint32(0)
		rmax uint32 = 0
		gmin uint32 = ^uint32(0)
		gmax uint32 = 0
		bmin uint32 = ^uint32(0)
		bmax uint32 = 0
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
		rsum uint32
		gsum uint32
		bsum uint32
	)
	l := uint32(len(arr))

	for _, p := range arr {
		rsum += p.r
		gsum += p.g
		bsum += p.b
	}

	return pix{
		r: rsum / l,
		g: gsum / l,
		b: bsum / l,
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
			flat[mX*j+i] = pix{r, g, b}
		}
	}

	return flat
}

func split(img_arr []pix, depth uint) pix {
	if len(img_arr) == 0 {
		return pix{0, 0, 0}
	}

	if depth == 0 {
		return average(img_arr)
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
		sort.SliceStable(img_arr, func(i, j int) bool { return img_arr[i].r < img_arr[i].r })
	case 1:
		sort.SliceStable(img_arr, func(i, j int) bool { return img_arr[i].g < img_arr[i].g })
	case 2:
		sort.SliceStable(img_arr, func(i, j int) bool { return img_arr[i].b < img_arr[i].b })
	}

	medianIndex := len(img_arr) / 2

	// TODO: Use shared ds instead of returns
	split(img_arr[:medianIndex], depth-1)
	return split(img_arr[medianIndex:], depth-1)
}

// returns hexcodes for palette of num colors where num is a power of 2
func GetPaletteMC(img *image.Image, n uint) ([]string, error) {
	if notPowerOfTwo(n) {
		err := errors.New(
			fmt.Sprintf("n must be power of two. Received %d.", n),
		)
		return []string{}, err
	}

	flat := flatten(*img)
	split(flat, n)
	return []string{}, nil
}
