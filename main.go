package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"os"

	"goPalettes/imageManip"
	"goPalettes/ui"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

var programState ui.State

func main() {
	filePath := flag.String("h", "", "File to use if running in headless mode.")
	flag.Parse()

	if *filePath != "" {
		runHeadless(*filePath)
		return
	}

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

func runHeadless(filePath string) {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file (%s): %s\n", filePath, err.Error())
		return
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Printf("Error decoding image (%s): %s\n", filePath, err.Error())
		return
	}

	colors := imageManip.GetPaletteMC(&img, 4)

	for _, c := range colors {
		fmt.Printf("%s\n", c)
	}
}
