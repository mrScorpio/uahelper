package ui

import (
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/mrscorpio/uahelper/configs"
	"github.com/mrscorpio/uahelper/internal/tagdata"
)

type Taglist struct {
	w        *app.Window
	pp       []widget.Bool
	list     layout.List
	isOpen   bool
	names    []string
	SelTags  []int
	ShowTags map[int]bool
}

func NewTaglist(d *tagdata.AllTags, cfg *configs.Config) *Taglist {
	w := new(app.Window)

	w.Option(app.Size(unit.Dp(222), unit.Dp(444)))
	//w.Option(app.Decorated(false))

	pp := make([]widget.Bool, len(d.Tag))
	for key, v := range cfg.ShowTags {
		pp[key].Value = v
	}

	names := make([]string, len(d.Tag))
	selTags := make([]int, 0)
	for i := range names {
		names[i] = d.Tag[i].Name
	}
	return &Taglist{
		w:        w,
		pp:       pp,
		list:     layout.List{Axis: layout.Vertical},
		names:    names,
		SelTags:  selTags,
		ShowTags: cfg.ShowTags,
	}
}

func (d *Taglist) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// Стиль кнопки выпадающего списка

	var widgets []layout.Widget

	// Если список открыт, добавляем элементы
	if d.isOpen {
		for i, item := range d.names {
			widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
				btn := material.CheckBox(th, &d.pp[i], item)
				btn.Color = th.Palette.Fg
				if d.pp[i].Pressed() {
					d.SelTags = append(d.SelTags, i)
					d.ShowTags[i] = !d.pp[i].Value
				}
				return btn.Layout(gtx)
			})
		}
	}

	return d.list.Layout(gtx, len(widgets), func(gtx C, index int) D {
		return widgets[index](gtx)
	})
}

func (d *Taglist) DrawPopup(th *material.Theme, cfg *configs.Config) error {
	var ops op.Ops
	d.w.Option(app.Size(unit.Dp(222), unit.Dp(444)))
	for {
		evt := d.w.Event()

		switch typ := evt.(type) {
		case app.FrameEvent:
			gtx := app.NewContext(&ops, typ)

			layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceStart,
			}.Layout(gtx,
				layout.Rigid(
					func(gtx C) D {
						return d.Layout(gtx, th)
					},
				))
			typ.Frame(gtx.Ops)
		case app.DestroyEvent:
			d.isOpen = false
			cfg.WrFile()
			return typ.Err

		}
	}
}
