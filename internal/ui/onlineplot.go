package ui

import (
	"image"

	"github.com/mrscorpio/uahelper/internal/tagdata"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

func DrawPlot(d *tagdata.AllTags) error {
	p := plot.New()
	last := len(d.Tag[1].Y)
	pts := make(plotter.XYs, 666)
	j := 0
	for i := last - 666; i < last; i++ {
		pts[j].X = float64(j)
		pts[j].Y = float64(d.Tag[1].Y[i].Value.(float32))
		j++
	}
	err := plotutil.AddLinePoints(p, pts)
	if err != nil {
		return err
	}
	img := image.NewRGBA(image.Rect(0, 0, 400, 300))
	c := vgimg.NewWith(vgimg.UseImage(img))
	p.Draw(draw.New(c))
	BufImg = c.Image()

	return nil
}
