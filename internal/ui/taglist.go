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
	dscrs    []string
	SelTags  []string
	ShowTags map[int]bool
	ShowHint []bool
}

func NewTaglist(d *tagdata.AllTags, cfg *configs.Config) *Taglist {
	w := new(app.Window)

	//w.Option(app.Decorated(false))
	for _, v := range cfg.ShowTagNames {
		for i, tag := range d.Tag {
			if v == tag.Name {
				cfg.ShowTags[i] = true
				break
			}
		}
	}

	pp := make([]widget.Bool, len(d.Tag))
	for key, v := range cfg.ShowTags {
		pp[key].Value = v
	}

	names := make([]string, len(d.Tag))
	dscrs := make([]string, len(d.Tag))
	showHint := make([]bool, len(d.Tag))
	for i := range names {
		names[i] = d.Tag[i].Name
		dscrs[i] = " " + d.Tag[i].Dscr + ", " + d.Tag[i].Unit
	}
	return &Taglist{
		w:        w,
		pp:       pp,
		list:     layout.List{Axis: layout.Vertical},
		names:    names,
		dscrs:    dscrs,
		SelTags:  cfg.ShowTagNames,
		ShowTags: cfg.ShowTags,
		ShowHint: showHint,
	}
}

func (tl *Taglist) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// Стиль кнопки выпадающего списка

	var widgets []layout.Widget

	// Если список открыт, добавляем элементы
	if tl.isOpen {
		for i, item := range tl.names {
			widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
				btn := material.CheckBox(th, &tl.pp[i], item)
				btn.Color = th.Palette.Fg
				if tl.pp[i].Update(gtx) {
					tl.ShowTags[i] = tl.pp[i].Value
					if !tl.ShowTags[i] {
						delete(tl.ShowTags, i)
					}
					tl.SelTags = make([]string, 0)
					for k := range tl.ShowTags {
						tl.SelTags = append(tl.SelTags, tl.names[k])
					}
				}
				if tl.pp[i].Hovered() {
					btn.Label = item + tl.dscrs[i]
				}
				return btn.Layout(gtx)
			})
			/*
				if tl.ShowHint[i] {
					widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
						hint := material.Label(th, unit.Sp(6), tl.dscrs[i])
						return hint.Layout(gtx)
					})
				}
			*/
		}
	}

	return tl.list.Layout(gtx, len(widgets), func(gtx C, index int) D {
		return widgets[index](gtx)
	})
}

func (tl *Taglist) DrawPopup(th *material.Theme, cfg *configs.Config) error {
	var ops op.Ops
	tl.w.Option(app.Size(unit.Dp(300), unit.Dp(500)))
	tl.w.Option(app.Title("Список тэгов"))
	for {
		evt := tl.w.Event()

		switch typ := evt.(type) {
		case app.FrameEvent:
			gtx := app.NewContext(&ops, typ)

			layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceStart,
			}.Layout(gtx,
				layout.Rigid(
					func(gtx C) D {
						return tl.Layout(gtx, th)
					},
				))
			typ.Frame(gtx.Ops)
		case app.DestroyEvent:
			tl.isOpen = false
			for key, v := range cfg.ShowTags {
				if !v {
					delete(cfg.ShowTags, key)
				}
			}
			cfg.ShowTagNames = tl.SelTags
			cfg.WrFile()
			return typ.Err

		}
	}
}

func (tl *Taglist) UpdTaglist(d *tagdata.AllTags, cfg *configs.Config) {
	for _, v := range cfg.ShowTagNames {
		for i, tag := range d.Tag {
			if v == tag.Name {
				cfg.ShowTags[i] = true
				break
			}
		}
	}

	tl.pp = make([]widget.Bool, len(d.Tag))
	for key, v := range cfg.ShowTags {
		tl.pp[key].Value = v
	}

	tl.names = make([]string, len(d.Tag))
	for i := range tl.names {
		tl.names[i] = d.Tag[i].Name
	}
}
