package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/gopcua/opcua"
	"github.com/mrscorpio/uahelper/configs"
	"github.com/mrscorpio/uahelper/internal/repository"
	"github.com/mrscorpio/uahelper/internal/tagdata"
	"github.com/mrscorpio/uahelper/pkg/opcuacl"
	"github.com/mrscorpio/uahelper/pkg/tgbot"

	"github.com/go-telegram/bot"
)

const (
	MdRd bool = false
)

var (
	d      tagdata.AllTags
	legSel map[string]bool
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := configs.LoadConfig()
	arhDirName := "arh/"

	fmt.Println("тренды пялить на localhost" + cfg.TrPort + "/?zoom=tag_for_right_axis&show=tag1,tag2,...")

	b, err := tgbot.NewBot(cfg.BotToken, MdRd || !cfg.Bot)
	if err != nil {
		log.Println(err)
	}

	cl, err := opcuacl.NewCl(ctx, cfg, MdRd)
	if err != nil {
		log.Fatal(err)
	}

	legSel = make(map[string]bool)

	var wg sync.WaitGroup

	if !MdRd {
		fmt.Println("читаем сервачок", cfg.Endpoint)
		os.Mkdir(strings.TrimSuffix(arhDirName, "/"), 0755)

		filedata, err := os.ReadFile(arhDirName + time.Now().Format("20060102_15") + ".json")
		if err == nil {
			err := json.Unmarshal(filedata, &d)
			if err != nil {
				log.Println(err)
			}
		}

		err = d.ReadOpcTagList(ctx, cl)
		if err != nil {
			log.Println(err)
		}

		spin := false
		rpmInd := 0
		for i := range d.Tag {
			if d.Tag[i].Name == "ST50_BZK" {
				rpmInd = i
			}
		}

		if cfg.Bot {
			go b.Start(ctx)
		}
		wg.Add(1)
		go func() {
			if b != nil {
				b.SendMessage(ctx, &bot.SendMessageParams{ChatID: -1003556463783, Text: "логер запущен"})
			}
			defer wg.Done()
			ticker := time.NewTicker(time.Duration(cfg.StoreCycle) * time.Second)
			chkSpin := time.NewTicker(6 * time.Second)
			current_hour := time.Now().Hour()

			for {
				select {
				case <-ctx.Done():
					log.Println("data process stopped")
					return
				case <-chkSpin.C:
					var curRpm float32
					if len(d.Tag[rpmInd].Y) > 0 {
						curRpm = d.Tag[rpmInd].Y[len(d.Tag[rpmInd].Y)-1].Value.(float32)
					}

					if !spin && curRpm > 6.6 {
						spin = true
						d.Clean()
					}
					if spin && curRpm < 6.6 {
						spin = false
						err := repository.StoreData(&d, arhDirName, false, cfg, b, ctx)
						if err != nil {
							log.Println(err)
						}
						if b != nil {
							b.SendMessage(ctx, &bot.SendMessageParams{ChatID: -1003556463783, Text: "остановились"})
						}
					}

				case <-ticker.C:
					nowT := time.Now()

					if nowT.Hour() != current_hour && !spin {
						current_hour = nowT.Hour()

						err := repository.StoreData(&d, arhDirName, true, cfg, b, ctx)
						if err != nil {
							log.Println(err)
						} else {
							d.Clean()
						}
					} else {
						data, err := json.Marshal(d)
						if err != nil {
							log.Println(err)
						}
						err = os.WriteFile(arhDirName+nowT.Format("20060102_15")+".json", data, 0755)
						if err != nil {
							log.Println(err)
						}

					}

				default:
					newTm := ""
					crTm := ""
					for key, item := range d.Ccs {
						if item.Cct >= key {
							if cl.State() == opcua.Connected {
								item.Resp, err = cl.Read(ctx, item.Req)
							}
							if err != nil {
								log.Fatal("opcua request error: ", err)
							}

							item.Cct = 0
						}

						for i := range item.Resp.Results {
							crTm = item.Resp.Results[i].ServerTimestamp.Local().Format("15:04:05.000")
							d.AddV(item.FirstPos+i, item.Resp.Results[i].Value.Value().(float32), crTm)
						}

						item.Cct += d.MinCycle

						if item.Cct <= d.MinCycle {
							newTm = crTm
						}
					}
					if newTm != "" {
						d.Tm = append(d.Tm, newTm)
					}
					time.Sleep(time.Duration(d.MinCycle) * time.Millisecond)
				}
			}
		}()
	}

	mux := http.NewServeMux()
	srv := http.Server{
		Addr:    cfg.TrPort,
		Handler: mux,
	}
	stopSrvSig := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-stopSrvSig
		err := srv.Shutdown(ctx)
		if err != nil {
			log.Println(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		mux.HandleFunc("/", trendView)
		err := srv.ListenAndServe()
		if err != nil {
			log.Println(err)
		}

	}()

	filename := ""

	for {
		if MdRd {
			fmt.Printf("для останова введи ку\nчто именно пялим > ")
		} else {
			fmt.Print("для останова введи ку > ")
		}
		fmt.Scan(&filename)
		if strings.TrimSpace(filename) == "q" {
			break
		}
		err := repository.ReadStored(&d, filename)
		if err != nil {
			log.Println(err)
			continue
		}

		fmt.Println("загружено, смотри в браузере")
	}
	if b != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: -1003556463783, Text: "логер остановлен"})
	}
	close(stopSrvSig)
	if cl != nil {
		cl.Close(ctx)
	}
	cancel()

	wg.Wait()
}

func trendView(w http.ResponseWriter, req *http.Request) {

	line := charts.NewLine()

	chsdTags := strings.Split(req.URL.Query().Get("show"), ",")
	tag2 := strings.Split(req.URL.Query().Get("zoom"), ",")

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

			line.SetXAxis(d.Tm)
			seriesName := d.Tag[v].Name + "_" + d.Tag[v].Dscr
			line.AddSeries(seriesName, d.Tag[v].Y,
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

	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme:     types.ThemeWesteros,
			Width:     "1777px",
			Height:    "888px",
			PageTitle: "чёткие трендики",
		}),
		charts.WithGridOpts(opts.Grid{Width: "999px"}),
		charts.WithLegendOpts(opts.Legend{Type: "scroll", Orient: "vertical", X: "right", Selected: legSel}),
		charts.WithDataZoomOpts(
			opts.DataZoom{Type: "slider", Orient: "horizontal"},
			opts.DataZoom{Type: "inside", Orient: "vertical", YAxisIndex: zoomAxis},
		),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
	)

	line.Render(w)

}
