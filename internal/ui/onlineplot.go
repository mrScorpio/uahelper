package ui

import (
	"image"

	"github.com/mrscorpio/uahelper/internal/tagdata"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/text"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

func DrawPlot(d *tagdata.AllTags) error {
	p := plot.New()
	last := len(d.Tag[1].Y)
	pts := make(plotter.XYs, 66)
	j := 0
	for i := last - 66; i < last; i++ {
		pts[j].X = float64(j)
		pts[j].Y = d.Tag[1].Y[i].Value.(float64)
		j++
	}
	err := plotutil.AddLinePoints(p, pts)
	p.X.Tick.Label.Rotation = 3.14 / 2
	p.X.Tick.Label.XAlign = text.XRight
	p.Y.Max = d.Tag[1].Max
	p.Y.Min = d.Tag[1].Min
	p.Y.Label.Text = "%"
	p.Y.Label.Position = text.PosTop
	p.Y.Label.TextStyle.Rotation = 3.14 / 2

	if err != nil {
		return err
	}
	img := image.NewRGBA(image.Rect(0, 0, 400, 300))
	c := vgimg.NewWith(vgimg.UseImage(img))
	p.Draw(draw.New(c))
	BufImg = c.Image()

	return nil
}
