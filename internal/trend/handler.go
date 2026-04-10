package trend

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/event"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/mrscorpio/uahelper/internal/tagdata"
)

// хэндлер для отрисовки трендов
func View(d *tagdata.AllTags, legSel map[string]bool, wTime *time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		line := charts.NewLine()

		chsdTags := strings.Split(req.URL.Query().Get("show"), ",")
		tag2 := strings.Split(req.URL.Query().Get("zoom"), ",")
		step, err := strconv.Atoi(req.URL.Query().Get("step"))
		if err != nil {
			step = 1
		}

		cnt := 1
		lcnt := 0
		axisW := 66
		zoomAxis := 1
		cmpUnit := ""
		var newAxis *opts.YAxis

		for key, item := range d.Unit {

			newAxis = &opts.YAxis{
				Name: key,
				//Min:      item.Min,
				//Max:      item.Max,
				Position:     "left",
				NameGap:      -lcnt * axisW,
				NameLocation: "middle",
				Scale:        opts.Bool(true),
				AlignTicks:   opts.Bool(true),
				AxisLine: &opts.AxisLine{
					OnZero: opts.Bool(false),
					LineStyle: &opts.LineStyle{
						Color: opts.RGBColor(uint16(lcnt*10), uint16(lcnt*20), uint16(lcnt*5)),
					},
				},
				AxisLabel: &opts.AxisLabel{
					Margin: -float64(lcnt * axisW),
					Color:  opts.RGBColor(uint16(lcnt*10), uint16(lcnt*20), uint16(lcnt*5)),
				},
			}

			if len(tag2) > 0 {
				tagname := strings.ToUpper(strings.TrimSpace(tag2[0]))

				for _, v := range item.Pos {
					if tagname == d.Tag[v].Name {
						cmpUnit = key
						chsdTags = append(chsdTags, tag2[0])
						break
					}
				}
			}

			if key == cmpUnit {
				zoomAxis = cnt
				lcnt--
				newAxis.Position = "right"
				newAxis.NameGap = -33
				newAxis.AxisLabel = &opts.AxisLabel{Margin: -33.3}
			}

			line.ExtendYAxis(*newAxis)

			for _, v := range item.Pos {
				lenShow := len(d.Tag[v].Y) / step

				tmShow := make([]string, lenShow)
				vShow := make([]opts.LineData, lenShow)

				for i := range tmShow { //прореживание
					tmShow[i] = d.Tm[i*step]
					vShow[i] = d.Tag[v].Y[i*step]
				}

				line.SetXAxis(tmShow)

				seriesName := d.Tag[v].Name + "_" + d.Tag[v].Dscr

				line.AddSeries(seriesName, vShow,
					charts.WithDatasetIndex(v),
					charts.WithLineChartOpts(opts.LineChart{YAxisIndex: cnt}),
				)

				legSel[seriesName] = false
			}
			cnt++
			lcnt++
		}

		for _, v := range chsdTags {
			tagname := strings.ToUpper(strings.TrimSpace(v))
			legSel[tagname+"_"+d.Descr[tagname]] = true
		}

		clickHandler := `(params) => alert(params.seriesIndex)`

		line.SetGlobalOptions(
			charts.WithInitializationOpts(opts.Initialization{
				Theme:     types.ThemeWesteros,
				Width:     "1777px",
				Height:    "888px",
				PageTitle: "чёткие трендики",
			}),
			charts.WithTitleOpts(opts.Title{Title: wTime.Format(time.Stamp), Left: "center"}),
			charts.WithGridOpts(opts.Grid{Width: "999px"}),
			charts.WithLegendOpts(opts.Legend{Type: "scroll", Orient: "vertical", X: "right", Selected: legSel}),
			charts.WithDataZoomOpts(
				opts.DataZoom{Type: "slider", Orient: "horizontal"},
				opts.DataZoom{Type: "inside", Orient: "vertical", YAxisIndex: zoomAxis},
			),
			charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
			charts.WithEventListeners(
				event.Listener{
					EventName: "dblclick",
					Handler:   opts.FuncOpts(clickHandler),
				},
			),
		)

		line.Render(w)

	}
}
