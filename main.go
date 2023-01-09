package main

import (
	"log"
	"os"

	"goPalettes/ui"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

var programState ui.State

func main() {

	go func() {
		w := app.NewWindow(
			app.Title("goPalettes"),
			app.MinSize(unit.Dp(300), unit.Dp(600)),
		)

		programState.Init(w)
		if err := draw(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func draw(w *app.Window) error {
	var ops op.Ops

	for e := range w.Events() {
		switch e := e.(type) {
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			programState.Layout(gtx)
			e.Frame(gtx.Ops)
		case system.DestroyEvent:
			return e.Err
		}
	}
	return nil
}
