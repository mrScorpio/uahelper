package ui

import (
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type Dropdown struct {
	items      []string
	selected   int
	isOpen     bool
	list       layout.List
	dropdown   widget.Clickable
	ppw        *app.Window
	itemClicks []widget.Clickable
}

func NewDropdown(items []string) *Dropdown {
	itemClicks := make([]widget.Clickable, len(items))

	ppw := new(app.Window)

	return &Dropdown{
		items:      items,
		selected:   -1, // ничего не выбрано
		list:       layout.List{Axis: layout.Vertical},
		itemClicks: itemClicks,
		ppw:        ppw,
	}
}

func (d *Dropdown) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// Стиль кнопки выпадающего списка
	button := material.Button(th, &d.dropdown, d.getSelectedText())
	button.Background = th.Palette.Bg
	button.Color = th.Palette.Fg

	if d.dropdown.Clicked(gtx) {
		d.isOpen = !d.isOpen
		/*
			d.ppw.Option(app.Size(unit.Dp(200), unit.Dp(300)))
			d.ppw.Option(app.Decorated(false))
			if d.isOpen {
				go d.DrawPopup(d.ppw, th)
			} else {
				d.ppw.Invalidate()
				d.ppw.Perform(system.ActionClose)
			}
		*/
	}

	var widgets []layout.Widget

	// Добавляем кнопку как первый элемент
	widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
		return button.Layout(gtx)
	})

	// Если список открыт, добавляем элементы
	if d.isOpen {
		for i, item := range d.items {
			itemIndex := i
			widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(th, &d.itemClicks[i], item)
				btn.Background = th.Palette.Bg
				btn.Color = th.Palette.Fg
				if d.itemClicks[i].Clicked(gtx) {
					d.selected = itemIndex
					d.isOpen = false
					NewData <- d.items[d.selected]
				}
				return btn.Layout(gtx)
			})
		}
	}

	return d.list.Layout(gtx, len(widgets), func(gtx C, index int) D {
		return widgets[index](gtx)
	})
}

func (d *Dropdown) getSelectedText() string {
	if d.selected == -1 {
		return "Выберите файл"
	}
	return d.items[d.selected]
}

func (d *Dropdown) DrawPopup(w *app.Window, th *material.Theme) error {
	for {
		var ops op.Ops
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
						return d.Layout(gtx, th)
					},
				))
			typ.Frame(gtx.Ops)
		case app.DestroyEvent:
			return typ.Err

		}
	}
}
