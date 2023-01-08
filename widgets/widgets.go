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
	Th               *material.Theme
	CurImg           image.Image
	CurImgWidget     widget.Image
	Palette          []ColorBlock
	LoadingPalette   bool
	ButtonGetPalette widget.Clickable
}

func (s *State) Init() {
	s.Th = material.NewTheme(gofont.Collection())
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

	s.CurImg = img
	s.CurImgWidget.Src = paint.NewImageOp(img)
	s.CurImgWidget.Fit = widget.ScaleDown
	s.CurImgWidget.Position = layout.Center

	return nil
}

func (s *State) Layout(w *app.Window, gtx C) {
	if s.ButtonGetPalette.Clicked() {
		go func() {
			s.LoadingPalette = true
			numOfColors := 5
			goRoutines := runtime.NumCPU()
			var tolerance float64 = 10
			colors := imageManip.ExtractPaletteConcurrent(
				s.CurImg,
				numOfColors,
				goRoutines,
				tolerance,
			)
			fmt.Println(colors)
			p := make([]ColorBlock, len(colors))
			for i, c := range colors {
				p[i] = CreateColorBlock(c.ColString)
			}
			s.Palette = p
			s.LoadingPalette = false
			w.Invalidate()
		}()
	}

	layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceStart,
	}.Layout(gtx,
		layout.Flexed(1,
			ImageSection(gtx, &s.CurImgWidget),
		),
		layout.Rigid(
			PaletteSection(gtx, s),
		),
		layout.Rigid(
			ButtonGetPalette(gtx, s),
		),
	)

}

func ImageSection(gtx C, imgWidget *widget.Image) layout.Widget {
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

func PaletteSection(gtx C, s *State) layout.Widget {
	margins := layout.Inset{
		Left: unit.Dp(MARGIN1),
	}

	label := material.H3(s.Th, "Palette: ").Layout
	var innerWidget layout.Widget
	if s.LoadingPalette {
		innerWidget = material.H6(s.Th, "Generating palette...").Layout
	} else if len(s.Palette) == 0 {
		innerWidget = material.H6(s.Th, "None").Layout
	} else {
		var children []layout.FlexChild
		for i := range s.Palette {
			block := &s.Palette[i]
			children = append(children,
				layout.Rigid(func(gtx C) D {
					return layout.UniformInset(unit.Dp(10)).Layout(gtx, block.Layout)
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

type ColorBlock struct {
	hexCode string
	col     color.NRGBA
}

func CreateColorBlock(hexCode string) ColorBlock {
	return ColorBlock{
		hexCode: hexCode,
		col:     imageManip.HexToNRGBA(hexCode),
	}
}

func (c *ColorBlock) Layout(gtx C) D {
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

func ButtonGetPalette(gtx C, s *State) layout.Widget {
	margins := layout.UniformInset(unit.Dp(MARGIN1))

	return func(gtx C) D {
		return margins.Layout(gtx,
			func(gtx C) D {
				btn := material.Button(s.Th, &s.ButtonGetPalette, "Get palette")
				return btn.Layout(gtx)
			},
		)
	}
}
