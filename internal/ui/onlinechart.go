package ui

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"strconv"
	"strings"

	"github.com/mrscorpio/uahelper/configs"
	"github.com/mrscorpio/uahelper/internal/repository"
	"github.com/mrscorpio/uahelper/pkg/tagdata"
	vcharts "github.com/vicanso/go-charts/v2"
)

func DrawChart(cfg *configs.Config) error {
	Mu.RLock()
	d := TrendData[*PrevHour]
	Mu.RUnlock()
	values := make([][]float64, len(cfg.ShowTags))
	//legend := make([]string, 0)
	i := 0
	//order := make(map[int]int)
	Mu.Lock()
	if len(TagOrder) != len(cfg.ShowTags) {
		TagOrder = make(map[int]int)
		TagLegend = make([]string, 0)
	}
	Mu.Unlock()
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

	numPoints := ChartW

	for key, v := range cfg.ShowTags {
		if v {
			values[i] = make([]float64, numPoints)
			if len(d.Tag[key].V) < 6 {
				d.Mu.RUnlock()

				Cmd <- 1
				return nil
			}
			Mu.Lock()
			if len(TagOrder) != len(cfg.ShowTags) {
				TagOrder[key] = i
				TagLegend = append(TagLegend, d.Tag[key].Name)
			}
			Mu.Unlock()
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

	tm := make([]string, numPoints)
	j := LastInd
	//if j > 0 {
	//prev := make(map[int]float64)
	cfg.Mu.Lock()
	for i := numPoints - 1; i >= 0; i-- {

		for key, v := range cfg.ShowTags {
			if v {
				if j > int64(len(d.Tag[key].V)-1) {
					j = int64(len(d.Tag[key].V) - 1)
				}
				Mu.RLock()
				values[TagOrder[key]][i] = float64(d.Tag[key].V[j])
				Mu.RUnlock()
			}
		}
		tm[i] = d.Tt[j].Format("15:04:05")
		if j > FstInd {
			j -= Diap / int64(numPoints)
		}
		if j < 0 {
			j = 0
		}
	}
	cfg.Mu.Unlock()
	//}
	d.Mu.RUnlock()

	scMax := float64(ScMax)
	scMin := float64(ScMin)

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
			opt.YAxisOptions[0].Max = &scMax
			opt.YAxisOptions[0].Min = &scMin
			//opt.SeriesList[0].Label = vcharts.SeriesLabel{Formatter: "{c}", Show: *vcharts.TrueFlag()}
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

func ShowHour(cfg *configs.Config) (bool, error) {
	fl, err := os.ReadDir("arh")
	gogo := true
	if err != nil {
		return gogo, err
	}

	for i := range fl {
		inf, err := fl[len(fl)-1-i].Info()
		if err != nil {
			continue
		}
		if *PrevHour == 0 {
			break
		}
		if !strings.HasSuffix(inf.Name(), ".json") {
			res := strings.Split(inf.Name(), "_")
			if len(res) != 2 {
				continue
			}
			fileHour, err := strconv.Atoi(res[1])
			if err != nil {
				continue
			}
			if *PrevHour == -1 {
				FstHour = fileHour
			}
			if fileHour == FstHour+1+*PrevHour {
				if TrendData[*PrevHour] == nil {
					Mu.Lock()
					TrendData[*PrevHour] = new(tagdata.AllTags)
					_, err := repository.ReadStored(TrendData[*PrevHour], "arh/"+inf.Name())
					Mu.Unlock()
					if err != nil {
						return gogo, err
					}
				}
				err = DrawChart(cfg)
				if err != nil {
					return gogo, err
				}
				gogo = false

				return gogo, nil
			}
		}
	}
	*PrevHour = 0
	return gogo, nil
}
