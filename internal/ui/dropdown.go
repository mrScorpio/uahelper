package ui

import (
	"gioui.org/app"
	"gioui.org/layout"
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

func (dd *Dropdown) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// Стиль кнопки выпадающего списка
	button := material.Button(th, &dd.dropdown, dd.getSelectedText())
	button.Background = th.Palette.Bg
	button.Color = th.Palette.Fg

	if dd.dropdown.Clicked(gtx) {
		dd.isOpen = !dd.isOpen
	}

	var widgets []layout.Widget

	// Добавляем кнопку как первый элемент
	widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
		return button.Layout(gtx)
	})

	// Если список открыт, добавляем элементы
	if dd.isOpen {
		for i, item := range dd.items {
			itemIndex := i
			widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(th, &dd.itemClicks[i], item)
				btn.Background = th.Palette.Bg
				btn.Color = th.Palette.Fg
				if dd.itemClicks[i].Clicked(gtx) {
					dd.selected = itemIndex
					dd.isOpen = false
					NewData <- dd.items[dd.selected]
				}
				return btn.Layout(gtx)
			})
		}
	}

	return dd.list.Layout(gtx, len(widgets), func(gtx C, index int) D {
		return widgets[index](gtx)
	})
}

func (dd *Dropdown) getSelectedText() string {
	if dd.selected == -1 {
		return "Выберите файл"
	}
	return dd.items[dd.selected]
}

/*
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
*/
