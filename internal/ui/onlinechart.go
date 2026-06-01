package ui

import (
	"bytes"
	"fmt"
	"image"

	"github.com/mrscorpio/uahelper/configs"
	"github.com/mrscorpio/uahelper/internal/tagdata"
	vcharts "github.com/vicanso/go-charts/v2"
)

func DrawChart(d *tagdata.AllTags, cfg *configs.Config) error {
	values := make([][]float64, len(cfg.ShowTags))
	//legend := make([]string, 0)
	i := 0
	//order := make(map[int]int)

	if len(TagOrder) != len(cfg.ShowTags) {
		TagOrder = make(map[int]int)
		TagLegend = make([]string, 0)
	}
	minInd := len(d.Tm) - 1
	for key, v := range cfg.ShowTags {
		if v {
			values[i] = make([]float64, Diap)
			if len(TagOrder) != len(cfg.ShowTags) {
				TagOrder[key] = i
				TagLegend = append(TagLegend, d.Tag[key].Name)
			}
			if len(d.Tag[key].Y) < minInd+1 {
				minInd = len(d.Tag[key].Y) - 1
			}
			if ScAuto {
				if d.Tag[key].Max > ScMax {
					ScMax = d.Tag[key].Max
				}
				if d.Tag[key].Min < ScMin || ScMin == 0.0 {
					ScMin = d.Tag[key].Min
				}
			}
			i++
		}
	}

	tm := make([]string, Diap)
	if Gogo {
		LastInd = minInd
	}
	j := LastInd
	if j > 0 {
		for i := Diap - 1; i >= 0; i-- {
			tm[i] = d.Tm[j]
			for key, v := range cfg.ShowTags {
				if v {
					values[TagOrder[key]][i] = d.Tag[key].Y[j].Value.(float64)
				}
			}
			if j > 0 {
				j--
			}
		}
	}

	p, err := vcharts.LineRender(
		values,
		vcharts.XAxisDataOptionFunc(tm),
		vcharts.LegendLabelsOptionFunc(TagLegend),
		vcharts.YAxisOptionFunc(vcharts.YAxisOption{Show: vcharts.TrueFlag()}),
		func(opt *vcharts.ChartOption) {
			opt.SymbolShow = vcharts.FalseFlag()
			opt.LineStrokeWidth = 1
			//opt.Title.Text = d.Tag[s].Unit
			opt.XAxis.FontSize = 8
			opt.YAxisOptions[0].FontSize = 8

			opt.Height = ChartH
			opt.Width = ChartW
			opt.YAxisOptions[0].Max = &ScMax
			opt.YAxisOptions[0].Min = &ScMin
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
