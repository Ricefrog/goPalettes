package widgets

import (
	"fmt"
	"goPalettes/imageManip"
	"image"
	"image/color"
	"os"
	"runtime"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

const (
	MARGIN1 = 25
)

type State struct {
	th               *material.Theme
	curImg           image.Image
	curImgWidget     widget.Image
	palette          []colorBlock
	loadingPalette   bool
	buttonGetPalette widget.Clickable
}

func (s *State) Init() {
	s.th = material.NewTheme(gofont.Collection())
}

func (s *State) SetCurImage(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return err
	}

	s.curImg = img
	s.curImgWidget.Src = paint.NewImageOp(img)
	s.curImgWidget.Fit = widget.ScaleDown
	s.curImgWidget.Position = layout.Center

	return nil
}

func (s *State) Layout(w *app.Window, gtx C) {
	if s.buttonGetPalette.Clicked() {
		go func() {
			s.loadingPalette = true
			numOfColors := 5
			goRoutines := runtime.NumCPU()
			var tolerance float64 = 10
			colors := imageManip.ExtractPaletteConcurrent(
				s.curImg,
				numOfColors,
				goRoutines,
				tolerance,
			)
			fmt.Println(colors)
			p := make([]colorBlock, len(colors))
			for i, c := range colors {
				p[i] = createColorBlock(c.ColString)
			}
			s.palette = p
			s.loadingPalette = false
			w.Invalidate()
		}()
	}

	layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceStart,
	}.Layout(gtx,
		layout.Flexed(1,
			imageSection(gtx, &s.curImgWidget),
		),
		layout.Rigid(
			paletteSection(gtx, s),
		),
		layout.Rigid(
			buttonGetPalette(gtx, s),
		),
	)

}

func imageSection(gtx C, imgWidget *widget.Image) layout.Widget {
	return func(gtx C) D {

		margins := layout.UniformInset(unit.Dp(MARGIN1))

		border := widget.Border{
			Color: color.NRGBA{R: 255, A: 255},
			Width: unit.Dp(5),
		}

		return margins.Layout(gtx,
			func(gtx C) D {
				return border.Layout(gtx,
					func(gtx C) D { return margins.Layout(gtx, imgWidget.Layout) },
				)
			},
		)
	}
}

func paletteSection(gtx C, s *State) layout.Widget {
	margins := layout.Inset{
		Left: unit.Dp(MARGIN1),
	}

	label := material.H3(s.th, "Palette: ").Layout
	var innerWidget layout.Widget
	if s.loadingPalette {
		innerWidget = material.H6(s.th, "Generating palette...").Layout
	} else if len(s.palette) == 0 {
		innerWidget = material.H6(s.th, "None").Layout
	} else {
		var children []layout.FlexChild
		for i := range s.palette {
			block := &s.palette[i]
			children = append(children,
				layout.Rigid(func(gtx C) D {
					return layout.UniformInset(unit.Dp(10)).Layout(gtx, block.layout)
				}),
			)
		}
		innerWidget = func(gtx C) D { return layout.Flex{}.Layout(gtx, children...) }
	}

	return func(gtx C) D {
		return margins.Layout(gtx,
			func(gtx C) D {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Spacing:   layout.SpaceEnd,
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Rigid(label),
					layout.Rigid(innerWidget),
				)
			},
		)
	}
}

type colorBlock struct {
	hexCode string
	col     color.NRGBA
}

func createColorBlock(hexCode string) colorBlock {
	return colorBlock{
		hexCode: hexCode,
		col:     imageManip.HexToNRGBA(hexCode),
	}
}

func (c *colorBlock) layout(gtx C) D {
	const size = 30
	yOffset := 5 // TODO: figure out how to make this dynamic based on height of label
	//yOffset := (gtx.Constraints.Max.Y - size) / 2
	//fmt.Printf("%v %d\n", gtx.Constraints, yOffset)

	for _, e := range gtx.Events(c) {
		if e, ok := e.(pointer.Event); ok {
			if e.Type == pointer.Press {
				fmt.Printf("%s was clicked.\n", c.hexCode)
			}
		}
	}

	op.Offset(image.Point{Y: yOffset}).Add(gtx.Ops)
	area := clip.Rect{
		Max: image.Point{size, size},
	}.Push(gtx.Ops)
	pointer.InputOp{Tag: c, Types: pointer.Press}.Add(gtx.Ops)
	pointer.CursorPointer.Add(gtx.Ops)

	paint.ColorOp{Color: c.col}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	area.Pop()

	return layout.Dimensions{Size: image.Point{X: size, Y: size}}
}

func buttonGetPalette(gtx C, s *State) layout.Widget {
	margins := layout.UniformInset(unit.Dp(MARGIN1))

	return func(gtx C) D {
		return margins.Layout(gtx,
			func(gtx C) D {
				btn := material.Button(s.th, &s.buttonGetPalette, "Get palette")
				return btn.Layout(gtx)
			},
		)
	}
}
