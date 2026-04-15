package ui

import (
	"image"
	"image/color"
	"log"
	"strconv"
	"strings"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type C = layout.Context
type D = layout.Dimensions

var (
	Progress float32
	ProgInc  chan float32
	NewData  chan string
	Gogo     bool
	Dura     float64
	BufImg   image.Image
)

func DrawSetup(w *app.Window) error {
	var ops op.Ops

	myBtn := new(widget.Clickable)
	mySld := new(widget.Float)
	duraInp := new(widget.Editor)
	duraInp.SingleLine = true
	duraInp.Alignment = text.Middle
	/*
		file, _ := os.Open("logo.png")
		img, _, err := image.Decode(file)
		if err != nil {
			fmt.Print(err)
		}
		file.Close()
	*/
	BufImg = image.NewRGBA(image.Rect(0, 0, 400, 300))

	go func() {
		for v := range ProgInc {
			if v == 1 {
				w.Invalidate()
			}
			if v == 6 {
				w.Perform(system.ActionClose)
			}
		}
	}()

	arhFiles, err := SearchDataFiles()
	if err != nil {
		log.Println(err)
	}

	filesDD := NewDropdown(arhFiles)

	th := material.NewTheme()

	for {
		evt := w.Event()

		switch typ := evt.(type) {
		case app.FrameEvent:
			gtx := app.NewContext(&ops, typ)

			if myBtn.Clicked(gtx) {
				Gogo = !Gogo
				inpDura := strings.TrimSpace(duraInp.Text())
				Dura, _ = strconv.ParseFloat(inpDura, 64)

			}

			if mySld.Dragging() {
				Gogo = false
				Progress += <-ProgInc
			}

			layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceEnd,
			}.Layout(gtx,
				layout.Rigid(
					func(gtx C) D {
						margins := layout.Inset{
							Top:    unit.Dp(6),
							Bottom: unit.Dp(6),
							Left:   unit.Dp(6),
							Right:  unit.Dp(66),
						}
						brdr := widget.Border{
							Color: color.NRGBA{R: 6, G: 6, B: 6, A: 255},
							Width: unit.Dp(2),
						}
						ed := material.Editor(th, duraInp, "%")
						return margins.Layout(gtx,
							func(gtx C) D {
								return brdr.Layout(gtx, ed.Layout)
							},
						)
					},
				),

				layout.Rigid(
					func(gtx C) D {
						margins := layout.Inset{
							Top:    unit.Dp(19),
							Bottom: unit.Dp(19),
							Left:   unit.Dp(11),
							Right:  unit.Dp(11),
						}
						return margins.Layout(gtx,
							func(gtx C) D {

								/*
									circle := clip.Ellipse{
										// Hard coding the x coordinate. Try resizing the window
										// Min: image.Pt(80, 0),
										// Max: image.Pt(320, 240),
										// Soft coding the x coordinate. Try resizing the window
										Min: image.Pt(gtx.Constraints.Max.X/2-120, 0),
										Max: image.Pt(gtx.Constraints.Max.X/2+120, 240),
									}.Op(gtx.Ops)
									color := color.NRGBA{R: 200, A: 255}
									paint.FillShape(gtx.Ops, color, circle)
								*/
								return widget.Image{Src: paint.NewImageOp(BufImg), Fit: widget.Contain}.Layout(gtx)
							},
						)
					},
				),

				layout.Rigid(
					func(gtx C) D {
						margins := layout.Inset{
							Top:    unit.Dp(19),
							Bottom: unit.Dp(19),
							Left:   unit.Dp(11),
							Right:  unit.Dp(11),
						}
						return margins.Layout(gtx,
							func(gtx C) D {
								sld := material.Slider(th, mySld)
								return sld.Layout(gtx)
							},
						)
					},
				),
				layout.Rigid(
					func(gtx C) D {
						margins := layout.Inset{
							Top:    unit.Dp(6),
							Bottom: unit.Dp(6),
							Left:   unit.Dp(26),
							Right:  unit.Dp(26),
						}
						return margins.Layout(gtx,
							func(gtx C) D {
								btnTxt := "go"
								if Gogo {
									btnTxt = "stop"
								}
								btn := material.Button(th, myBtn, btnTxt)
								return btn.Layout(gtx)
							},
						)
					},
				),
				layout.Rigid(
					func(gtx C) D {
						margins := layout.Inset{
							Top:    unit.Dp(16),
							Bottom: unit.Dp(16),
							Left:   unit.Dp(16),
							Right:  unit.Dp(16),
						}
						return margins.Layout(gtx,
							func(gtx C) D {
								bar := material.ProgressBar(th, Progress)
								return bar.Layout(gtx)
							},
						)
					},
				),
				layout.Rigid(
					func(gtx C) D {
						margins := layout.Inset{
							Top:    unit.Dp(6),
							Bottom: unit.Dp(6),
							Left:   unit.Dp(6),
							Right:  unit.Dp(6),
						}
						return margins.Layout(gtx,
							func(gtx C) D {
								return layout.Flex{
									Axis:    layout.Horizontal,
									Spacing: layout.SpaceEnd,
								}.Layout(gtx,
									layout.Rigid(
										func(gtx C) D {
											margins := layout.Inset{
												Top:    unit.Dp(5),
												Bottom: unit.Dp(1),
												Left:   unit.Dp(3),
												Right:  unit.Dp(3),
											}
											return margins.Layout(gtx,
												func(gtx C) D {
													txt := material.H6(th, "Архив:")
													return txt.Layout(gtx)
												},
											)
										},
									),
									layout.Rigid(
										func(gtx C) D {
											brdr := widget.Border{
												Color: color.NRGBA{R: 6, G: 6, B: 222, A: 255},
												Width: unit.Dp(2),
											}
											return brdr.Layout(gtx,
												func(gtx C) D {
													return filesDD.Layout(gtx, th)
												},
											)

										},
									),
								)

							},
						)
					},
				),
			)

			typ.Frame(gtx.Ops)
		case app.DestroyEvent:
			return typ.Err
		}

	}
}
