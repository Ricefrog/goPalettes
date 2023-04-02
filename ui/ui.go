package ui

import (
	"fmt"
	"goPalettes/imageManip"
	"image"
	"image/color"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/clipboard"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/sqweek/dialog"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

const (
	MARGIN1 = 25
	MARGIN2 = 10
)

type State struct {
	th               *material.Theme
	curImg           image.Image
	curImgWidget     widget.Image
	palette          []colorBlock
	paletteMsg       string
	paletteMsgTimer  *time.Timer
	loadingPalette   bool
	buttonGetPalette widget.Clickable
	buttonChooseFile widget.Clickable
	w                *app.Window
}

func (s *State) Init(w *app.Window) {
	s.th = material.NewTheme(gofont.Collection())
	s.w = w
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

func (s *State) setPaletteMsg(msg string) {
	if s.paletteMsgTimer != nil {
		s.paletteMsgTimer.Stop()
	}

	s.paletteMsg = msg
	s.paletteMsgTimer = time.AfterFunc(time.Second*2, func() {
		s.paletteMsg = ""
		s.w.Invalidate()
	})
}

func (s *State) Layout(gtx C) {

	if s.buttonChooseFile.Clicked() {
		path, err := dialog.File().Filter("image", "png", "jpg").Load()
		if err != nil {
			if err != dialog.ErrCancelled {
				log.Fatal(err)
			}
		}

		if len(path) > 0 {
			err := s.SetCurImage(path)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	if s.buttonGetPalette.Clicked() {
		go func() {
			s.loadingPalette = true

			colors := imageManip.GetPaletteMC(&s.curImg, 4)
			p := make([]colorBlock, len(colors))
			for i, c := range colors {
				p[i] = createColorBlock(c)
			}

			s.palette = p
			s.loadingPalette = false
			s.w.Invalidate()
		}()
	}

	layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceStart,
	}.Layout(gtx,
		layout.Flexed(1,
			s.imageSection(gtx),
		),
		layout.Rigid(
			s.paletteSection(gtx),
		),
		layout.Rigid(
			s.controlPanelSection(gtx),
		),
	)

}

func (s *State) imageSection(gtx C) layout.Widget {
	return func(gtx C) D {

		margins := layout.UniformInset(unit.Dp(MARGIN1))

		border := widget.Border{
			Color: color.NRGBA{R: 255, A: 255},
			Width: unit.Dp(5),
		}

		var innerWidget layout.Widget
		if s.curImg == nil {
			innerWidget = material.H6(s.th, "No image selected.").Layout
		} else {
			innerWidget = s.curImgWidget.Layout
		}

		return margins.Layout(gtx,
			func(gtx C) D {
				return border.Layout(gtx,
					func(gtx C) D { return margins.Layout(gtx, innerWidget) },
				)
			},
		)
	}
}

func (s *State) paletteSection(gtx C) layout.Widget {
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
		margins2 := layout.UniformInset(MARGIN2)
		var children []layout.FlexChild
		for i := range s.palette {
			block := &s.palette[i]
			children = append(children,
				layout.Rigid(func(gtx C) D {
					return margins2.Layout(gtx, block.layout(gtx, s))
				}),
			)
		}
		children = append(children, layout.Rigid(
			func(gtx C) D {
				return margins2.Layout(gtx, material.H6(s.th, s.paletteMsg).Layout)
			},
		))

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

func (c *colorBlock) layout(gtx C, s *State) layout.Widget {
	border := widget.Border{
		Color: color.NRGBA{A: 255},
		Width: unit.Dp(2),
	}

	return func(gtx C) D {
		return border.Layout(gtx, func(gtx C) D {
			const size = 30
			yOffset := 5 // TODO: figure out how to make this dynamic based on height of label

			for _, e := range gtx.Events(c) {
				if e, ok := e.(pointer.Event); ok {
					if e.Type == pointer.Press {
						clipboard.WriteOp{Text: c.hexCode}.Add(gtx.Ops)
						go s.setPaletteMsg(fmt.Sprintf("Copied %s to clipboard.", c.hexCode))
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
		})
	}
}

func (s *State) controlPanelSection(gtx C) layout.Widget {
	margins := layout.UniformInset(unit.Dp(MARGIN1))

	return func(gtx C) D {
		return layout.Flex{
			Spacing: layout.SpaceEvenly,
		}.Layout(gtx,
			layout.Flexed(1, s.buttonWidget(gtx, "Get palette", &s.buttonGetPalette, margins, s.curImg == nil)),
			layout.Rigid(s.buttonWidget(gtx, "Choose file", &s.buttonChooseFile, margins, s.loadingPalette)),
		)
	}
}

func (s *State) buttonWidget(
	gtx C,
	label string,
	button *widget.Clickable,
	margins layout.Inset,
	disabled bool) layout.Widget {
	return func(gtx C) D {
		return margins.Layout(gtx,
			func(gtx C) D {
				if disabled {
					gtx = gtx.Disabled()
				}
				btn := material.Button(s.th, button, label)
				return btn.Layout(gtx)
			},
		)
	}
}
