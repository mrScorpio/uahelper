package ui

import (
	"bytes"
	"fmt"
	"image"

	"github.com/mrscorpio/uahelper/internal/tagdata"
	vcharts "github.com/vicanso/go-charts/v2"
)

func DrawChart(d *tagdata.AllTags) error {
	//last := len(d.Tag[1].Y) - 1
	values := make([][]float64, 6)
	values[0] = make([]float64, Diap)
	tm := make([]string, Diap)
	if Gogo {
		LastInd = len(d.Tag[1].Y) - 1
	}
	j := LastInd
	if j > 0 {
		for i := Diap - 1; i >= 0; i-- {
			tm[i] = d.Tm[j]
			values[0][i] = d.Tag[1].Y[j].Value.(float64)
			if j > 0 {
				j--
			}
		}
	}
	if ScAuto {
		if d.Tag[1].Max > ScMax {
			ScMax = d.Tag[1].Max
		}
		if d.Tag[1].Min < ScMin || ScMin == 0.0 {
			ScMin = d.Tag[1].Min
		}
	}
	p, err := vcharts.LineRender(
		values,
		vcharts.XAxisDataOptionFunc(tm),
		vcharts.YAxisOptionFunc(vcharts.YAxisOption{Min: &ScMin, Max: &ScMax}),
		func(opt *vcharts.ChartOption) {
			opt.SymbolShow = vcharts.FalseFlag()
			opt.LineStrokeWidth = 1
			opt.Title.Text = d.Tag[1].Unit
			opt.XAxis.FontSize = 8
			opt.YAxisOptions[0].FontSize = 8

			opt.Height = ChartH
			opt.Width = ChartW
			opt.ValueFormatter = func(f float64) string {
				return fmt.Sprintf("%.3f", f)
			}
		},
	)

	if err != nil {
		return err
	}

	BufChart, err := p.Bytes()
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	_, err = buf.Write(BufChart)
	if err != nil {
		return err
	}
	BufImg, _, err = image.Decode(&buf)
	if err != nil {
		return err
	}

	Cmd <- 1

	return nil
}
