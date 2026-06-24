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
	d.Mu.RLock()
	//minInd := len(d.Tt) - 1
	if Gogo {
		LastInd = int64(len(d.Tt) - 1)
	}
	Diap = LastInd - FstInd

	if Diap < 1 {
		d.Mu.RUnlock()
		BufImg = image.NewRGBA(image.Rect(0, 0, 22, 16))
		Cmd <- 1
		return nil
	}

	for key, v := range cfg.ShowTags {
		if v {
			values[i] = make([]float64, ChartW)
			if len(d.Tag[key].V) < 6 {
				d.Mu.RUnlock()

				Cmd <- 1
				return nil
			}
			if len(TagOrder) != len(cfg.ShowTags) {
				TagOrder[key] = i
				TagLegend = append(TagLegend, d.Tag[key].Name)
			}
			/*
				if len(d.Tag[key].V) < minInd+1 {
					minInd = len(d.Tag[key].V) - 1
				}
			*/
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

	tm := make([]string, ChartW)
	j := LastInd
	//if j > 0 {
	//prev := make(map[int]float64)
	cfg.Mu.Lock()
	for i := ChartW - 1; i >= 0; i-- {

		for key, v := range cfg.ShowTags {
			if v {
				if j > int64(len(d.Tag[key].V)-1) {
					j = int64(len(d.Tag[key].V) - 1)
				}
				values[TagOrder[key]][i] = d.Tag[key].V[j]
			}
		}
		tm[i] = d.Tt[j].Format("15:04:05")
		if j > FstInd {
			j -= Diap / int64(ChartW)
		}
		if j < 0 {
			j = 0
		}
	}
	cfg.Mu.Unlock()
	//}
	d.Mu.RUnlock()
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
