package main

import (
	"fmt"
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

type C = layout.Context
type D = layout.Dimensions

func main() {
	go func() {
		w := app.NewWindow(
			app.Title("goPalettes"),
		)

		if err := draw(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func draw(w *app.Window) error {
	var ops op.Ops
	var button widget.Clickable
	var imgWidget widget.Image

	f, err := os.Open("./images/basado1.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	img, format, err := image.Decode(f)
	fmt.Printf("format: %s\n", format)
	imgWidget.Src = paint.NewImageOp(img)
	imgWidget.Fit = widget.Contain

	th := material.NewTheme(gofont.Collection())
	for e := range w.Events() {
		switch e := e.(type) {
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)

			//viewBoxHeight := unit.Dp(gtx.Constraints.Max.Y * 3 / 4)
			//controlBoxHeight := gtx.Constraints.Max.Y / 4

			layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceStart,
			}.Layout(gtx,
				layout.Rigid(
					imageLayout(gtx, &imgWidget),
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

func buttonLayout(gtx C, button *widget.Clickable, th *material.Theme) layout.Widget {
	margins := layout.Inset{
		Top:    unit.Dp(25),
		Bottom: unit.Dp(25),
		Right:  unit.Dp(25),
		Left:   unit.Dp(25),
	}
	return func(gtx C) D {
		return margins.Layout(gtx,
			func(gtx C) D {
				btn := material.Button(th, button, "raise glass")
				return btn.Layout(gtx) // how can i pass in minimum dimension constraints??
				/*
					return btn.Layout(gtx,
						func(gtx C) D {
							return layout.Dimensions{
								Size: image.Point{Y: 50},
							}
						},
					)
				*/
			},
		)
	}
}

func imageLayout(gtx C, imgWidget *widget.Image) layout.Widget {
	return func(gtx C) D {
		margins := layout.Inset{
			Top:    unit.Dp(25),
			Bottom: unit.Dp(25),
			Right:  unit.Dp(25),
			Left:   unit.Dp(25),
		}

		border := widget.Border{
			Color: color.NRGBA{R: 255, A: 255},
			Width: unit.Dp(5),
		}

		return margins.Layout(gtx,
			func(gtx C) D {
				return border.Layout(gtx,
					imgWidget.Layout,
				)
			},
		)
	}
}
