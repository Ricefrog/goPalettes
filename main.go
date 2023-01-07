package main

import (
	"fmt"
	"goPalettes/imageManip"
	"image"
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type state struct {
	palette        []string
	loadingPalette bool
}

var programState state

func main() {
	f, err := os.Open("./images/basado1.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)

	go func() {
		w := app.NewWindow(
			app.Title("goPalettes"),
			app.MinSize(unit.Dp(300), unit.Dp(600)),
		)

		if err := draw(w, img); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func draw(w *app.Window, img image.Image) error {
	var ops op.Ops
	var button widget.Clickable
	var imgWidget widget.Image

	imgWidget.Src = paint.NewImageOp(img)
	imgWidget.Fit = widget.ScaleDown
	imgWidget.Position = layout.Center

	th := material.NewTheme(gofont.Collection())
	for e := range w.Events() {
		switch e := e.(type) {
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)

			if button.Clicked() {
				go func() {
					programState.loadingPalette = true
					numOfColors := 5
					var tolerance float64 = 10
					colors := imageManip.ExtractPalette(
						img,
						numOfColors,
						tolerance,
					)
					fmt.Println(colors)
					p := make([]string, len(colors))
					for i, c := range colors {
						p[i] = c.ColString
					}
					programState.palette = p
					programState.loadingPalette = false
					w.Invalidate()
				}()
			}

			layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceStart,
			}.Layout(gtx,
				layout.Flexed(1,
					imageLayout(gtx, &imgWidget),
				),
				layout.Rigid(
					paletteLayout(gtx, th),
				),
				layout.Rigid(
					buttonLayout(gtx, &button, th),
				),
			)

			e.Frame(gtx.Ops)
		case system.DestroyEvent:
			return e.Err
		}
	}
	return nil
}

func imageLayout(gtx C, imgWidget *widget.Image) layout.Widget {
	return func(gtx C) D {

		margins := layout.UniformInset(unit.Dp(25))

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

func paletteLayout(gtx C, th *material.Theme) layout.Widget {
	margins := layout.Inset{
		Left: unit.Dp(25),
	}

	label := material.H3(th, "Palette: ").Layout
	var innerWidget layout.Widget
	if programState.loadingPalette {
		innerWidget = material.H6(th, "Generating palette...").Layout
	} else if len(programState.palette) == 0 {
		innerWidget = material.H6(th, "None").Layout
	} else {
		innerWidget = material.H6(th, fmt.Sprintf("%v", programState.palette)).Layout
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

func buttonLayout(gtx C, button *widget.Clickable, th *material.Theme) layout.Widget {
	margins := layout.UniformInset(unit.Dp(25))

	return func(gtx C) D {
		return margins.Layout(gtx,
			func(gtx C) D {
				btn := material.Button(th, button, "Get palette")
				return btn.Layout(gtx) // how can i pass in minimum dimension constraints??
			},
		)
	}
}
