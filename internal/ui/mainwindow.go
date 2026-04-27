package ui

import (
	"image"
	"image/color"
	"log"
	"strconv"
	"strings"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/mrscorpio/uahelper/internal/tagdata"
)

type C = layout.Context
type D = layout.Dimensions

var (
	Diap    int64
	LastInd int
	Cmd     chan int
	NewData chan string
	Gogo    bool
	ScAuto  bool
	ScMax   float64
	ScMin   float64
	BufImg  image.Image
	ChartW  int
	ChartH  int
)

func DrawUi(w *app.Window, d *tagdata.AllTags) error {
	var ops op.Ops

	myBtn := new(widget.Clickable)

	swUpdPlot := new(widget.Bool)
	swUpdPlot.Value = true

	diapSld := new(widget.Float)
	Diap = 666
	diapSld.Value = float32(Diap) / 10000
	var diap float32 = 0.1

	crSld := new(widget.Float)
	crSld.Value = 1

	ScAuto = true
	maxInp := new(widget.Editor)
	maxInp.SingleLine = true
	maxInp.Alignment = text.Middle
	maxInp.Filter = "0123456789."
	maxInp.MaxLen = 6

	minInp := new(widget.Editor)
	minInp.SingleLine = true
	minInp.Alignment = text.Middle
	minInp.Filter = "0123456789."
	minInp.MaxLen = 6

	go func() {
		for v := range Cmd {

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
	var lastLen float32 = 0.0
	for {
		evt := w.Event()

		switch typ := evt.(type) {
		case app.FrameEvent:
			gtx := app.NewContext(&ops, typ)

			Gogo = swUpdPlot.Value
			diap = float32(len(d.Tm)) / float32(cap(d.Tm))

			if swUpdPlot.Pressed() {
				crSld.Value = 1
				lastLen = float32(len(d.Tm) - 1)
			}

			if crSld.Dragging() {
				if Gogo {
					lastLen = float32(len(d.Tm) - 1)
				}
				LastInd = int(crSld.Value * lastLen)
				if crSld.Value < 1 {
					swUpdPlot.Value = false
				}
				DrawChart(d)
			}

			//ops.Reset()
			event.Op(&ops, w)

			_, okp := typ.Source.Event(pointer.Filter{
				Target: maxInp,
				Kinds:  pointer.Leave,
			})

			_, okk := typ.Source.Event(key.Filter{
				Focus:    maxInp,
				Required: key.Modifiers(key.Press),
				Name:     key.NameEnter,
			})

			_, okp2 := typ.Source.Event(pointer.Filter{
				Target: minInp,
				Kinds:  pointer.Leave,
			})

			_, okk2 := typ.Source.Event(key.Filter{
				Focus:    minInp,
				Required: key.Modifiers(key.Press),
				Name:     key.NameEnter,
			})

			if okk || okp {
				//fmt.Print(".")
				inpMax := strings.TrimSpace(maxInp.Text())
				ScMax, _ = strconv.ParseFloat(inpMax, 64)

				if ScMax != 0.0 {
					ScAuto = false
					maxInp.SetCaret(0, 2)
				}
				if !Gogo {
					DrawChart(d)
				}
			}

			if okk2 || okp2 {
				inpMin := strings.TrimSpace(minInp.Text())
				ScMin, _ = strconv.ParseFloat(inpMin, 64)

				if ScMin != 0.0 {
					ScAuto = false
					minInp.SetCaret(0, 2)
				}
				if !Gogo {
					DrawChart(d)
				}
			}

			if diapSld.Dragging() {
				Diap = int64(diapSld.Value * 10000)
				ScAuto = true
				if !Gogo {
					DrawChart(d)
				}
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
							Right:  unit.Dp(6),
						}
						return margins.Layout(gtx,
							func(gtx C) D {
								return layout.Flex{
									Axis:    layout.Horizontal,
									Spacing: layout.SpaceBetween,
								}.Layout(gtx,
									layout.Rigid(
										func(gtx C) D {
											margins := layout.Inset{
												Top:    unit.Dp(1),
												Bottom: unit.Dp(1),
												Left:   unit.Dp(3),
												Right:  unit.Dp(3),
											}
											brdr := widget.Border{
												Color: color.NRGBA{R: 6, G: 6, B: 6, A: 255},
												Width: unit.Dp(2),
											}
											ed := material.Editor(th, maxInp, "  max  ")
											return brdr.Layout(gtx,
												func(gtx C) D {
													return margins.Layout(gtx, ed.Layout)
												},
											)
										},
									),
									layout.Flexed(6,
										func(gtx C) D {
											margins := layout.Inset{
												Top:    unit.Dp(1),
												Bottom: unit.Dp(1),
												Left:   unit.Dp(6),
												Right:  unit.Dp(6),
											}

											return margins.Layout(gtx,
												func(gtx C) D {
													sld := material.Slider(th, diapSld)

													return sld.Layout(gtx)
												},
											)
										},
									),
									layout.Rigid(
										func(gtx C) D {
											sw := material.Switch(th, swUpdPlot, "upd")
											return sw.Layout(gtx)
										},
									),
								)

							},
						)
					},
				),

				layout.Rigid(
					func(gtx C) D {
						margins := layout.Inset{
							Top:    unit.Dp(6),
							Bottom: unit.Dp(6),
							//Left:   unit.Dp(6),
							//Right:  unit.Dp(6),
						}
						ChartW = gtx.Constraints.Max.X
						ChartH = gtx.Constraints.Max.Y / 3 * 2
						return margins.Layout(gtx,
							func(gtx C) D {
								return widget.Image{Src: paint.NewImageOp(BufImg), Fit: widget.Contain}.Layout(gtx)
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
									Spacing: layout.SpaceBetween,
								}.Layout(gtx,
									layout.Rigid(
										func(gtx C) D {
											margins := layout.Inset{
												Top:    unit.Dp(1),
												Bottom: unit.Dp(1),
												Left:   unit.Dp(3),
												Right:  unit.Dp(3),
											}
											brdr := widget.Border{
												Color: color.NRGBA{R: 6, G: 6, B: 6, A: 255},
												Width: unit.Dp(2),
											}
											ed := material.Editor(th, minInp, "  min  ")
											return brdr.Layout(gtx,
												func(gtx C) D {
													return margins.Layout(gtx, ed.Layout)
												},
											)
										},
									),
									layout.Flexed(6,
										func(gtx C) D {
											margins := layout.Inset{
												Top:    unit.Dp(1),
												Bottom: unit.Dp(1),
												Left:   unit.Dp(6),
												Right:  unit.Dp(6),
											}

											return margins.Layout(gtx,
												func(gtx C) D {
													sld := material.Slider(th, crSld)

													return sld.Layout(gtx)
												},
											)
										},
									),
								)

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
								bar := material.ProgressBar(th, diap)
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
									layout.Rigid(
										func(gtx C) D {
											margins := layout.Inset{
												Top:    unit.Dp(1),
												Bottom: unit.Dp(1),
												Left:   unit.Dp(66),
												Right:  unit.Dp(6),
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
								)

							},
						)
					},
				),
			)

			typ.Frame(gtx.Ops)

		case app.DestroyEvent:
			return typ.Err
		case app.ConfigEvent:
			if !Gogo {
				DrawChart(d)
			}
		}

	}
}
