package ui

import (
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/mrscorpio/uahelper/configs"
	"github.com/mrscorpio/uahelper/internal/tagdata"
)

type Taglist struct {
	//	w        *app.Window
	pp    []widget.Bool
	list  layout.List
	isCls bool
	names []string
	dscrs []string
	//SelTags []string
	//ShowTags map[int]bool
	ShowHint []bool
	//Widgets  []layout.Widget
}

func NewTaglist(d *tagdata.AllTags, cfg *configs.Config) *Taglist {

	d.Mu.RLock()
	defer d.Mu.RUnlock()
	//w.Option(app.Decorated(false))
	for _, v := range cfg.ShowTagNames {
		for i, tag := range d.Tag {
			if v == tag.Name {
				cfg.Mu.RLock()
				cfg.ShowTags[i] = true
				cfg.Mu.RUnlock()
				break
			}
		}
	}

	pp := make([]widget.Bool, len(d.Tag))
	cfg.Mu.RLock()
	for key, v := range cfg.ShowTags {
		pp[key].Value = v
	}
	cfg.Mu.RUnlock()

	names := make([]string, len(d.Tag))
	dscrs := make([]string, len(d.Tag))

	for i := range names {
		names[i] = d.Tag[i].Name
		dscrs[i] = " " + d.Tag[i].Dscr + ", " + d.Tag[i].Unit
	}
	return &Taglist{
		pp:    pp,
		list:  layout.List{Axis: layout.Vertical},
		names: names,
		dscrs: dscrs,
		//SelTags: cfg.ShowTagNames,
		//ShowTags: cfg.ShowTags,
		isCls: true,
	}
}

func (tl *Taglist) Layout(gtx layout.Context, th *material.Theme, cfg *configs.Config) layout.Dimensions {
	// Если список открыт, добавляем элементы
	widgets := make([]layout.Widget, len(tl.names))
	if !tl.isCls {
		for i, item := range tl.names {
			widgets[i] = func(gtx layout.Context) layout.Dimensions {
				btn := material.CheckBox(th, &tl.pp[i], item)
				btn.Color = th.Palette.Fg
				if tl.pp[i].Update(gtx) {
					cfg.Mu.Lock()
					cfg.ShowTags[i] = tl.pp[i].Value
					if !cfg.ShowTags[i] {
						delete(cfg.ShowTags, i)
					}
					cfg.ShowTagNames = make([]string, 0)
					for k := range cfg.ShowTags {
						cfg.ShowTagNames = append(cfg.ShowTagNames, tl.names[k])
					}
					cfg.Mu.Unlock()
					Cmd <- 9
				}
				if tl.pp[i].Hovered() {
					btn.Label = item + tl.dscrs[i]
				}
				return btn.Layout(gtx)
			}
		}
	}

	return tl.list.Layout(gtx, len(widgets), func(gtx C, index int) D {
		return widgets[index](gtx)
	})
}

func (tl *Taglist) DrawPopup(w *app.Window, th *material.Theme, cfg *configs.Config) error {
	var ops op.Ops
	//defer cfg.WrFile()
	tl.isCls = false

	for {
		evt := w.Event()

		switch typ := evt.(type) {
		case app.FrameEvent:
			gtx := app.NewContext(&ops, typ)

			layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceStart,
			}.Layout(gtx,
				layout.Rigid(
					func(gtx C) D {
						return tl.Layout(gtx, th, cfg)
					},
				))
			typ.Frame(gtx.Ops)
		case app.DestroyEvent:
			/*
				for key, v := range cfg.ShowTags {
					if !v {
						delete(cfg.ShowTags, key)
					}
				}
			*/
			//cfg.ShowTagNames = tl.SelTags
			tl.isCls = true
			return typ.Err
		}
	}
}

func (tl *Taglist) UpdTaglist(d *tagdata.AllTags, cfg *configs.Config) {
	cfg.UpdTagMap(d)

	tl.pp = make([]widget.Bool, len(d.Tag))
	cfg.Mu.RLock()
	for key, v := range cfg.ShowTags {
		tl.pp[key].Value = v
	}
	cfg.Mu.RUnlock()

	tl.names = make([]string, len(d.Tag))
	tl.dscrs = make([]string, len(d.Tag))
	for i := range tl.names {
		tl.names[i] = d.Tag[i].Name
		tl.dscrs[i] = " " + d.Tag[i].Dscr + ", " + d.Tag[i].Unit
	}
	Diap = int64(len(d.Tt))
	LastInd = Diap
}
