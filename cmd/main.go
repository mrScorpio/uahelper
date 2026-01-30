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
	"github.com/gopcua/opcua/ua"
	"github.com/mrscorpio/uahelper/configs"
	"github.com/mrscorpio/uahelper/internal/repository"
	"github.com/mrscorpio/uahelper/internal/tagdata"
	"github.com/mrscorpio/uahelper/internal/tripreport"
	"github.com/mrscorpio/uahelper/pkg/opcuacl"
	"github.com/mrscorpio/uahelper/pkg/tgbot"
)

const (
	MdRd bool = true // для выбора перед компиляцией - логер 0 или просмотр 1
)

var (
	d      tagdata.AllTags // пока структура с данными - это глобальная переменная
	legSel map[string]bool // для отключения позиций легенды в трендах
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := configs.LoadConfig()

	arhDirName := "arh/" //папка для хранения файлов

	fmt.Println("тренды пялить на localhost" + cfg.TrPort + "/?zoom=tag_for_right_axis&show=tag1,tag2,...")

	b, err := tgbot.NewBot(ctx, cfg, MdRd) // бот, в режиме просмотра нил
	if err != nil {
		log.Println(err)
	}

	cl, err := opcuacl.NewCl(ctx, cfg, MdRd) // описи юа клиент, в режиме просмотра нил
	if err != nil {
		log.Fatal(err)
	}

	legSel = make(map[string]bool) // для отключения позиций легенды в трендах

	var wg sync.WaitGroup

	if !MdRd {
		fmt.Println("читаем сервачок", cfg.Endpoint)
		os.Mkdir(strings.TrimSuffix(arhDirName, "/"), 0755)

		var prevFirst uint32 = 6 // чтобы не спамить авариями

		// тут тэги с первопричинами аварии
		tagname := []string{
			"PROT.FIRSTCOM.FIRST",
			"PROT.TRIP.FIRST",
		}
		tripTags := make([]*ua.ReadValueID, 0)
		for _, v := range tagname {
			id, err := ua.ParseNodeID("ns=1;s=REGUL_R500." + v)
			if err != nil {
				fmt.Println(err)
			}
			tripTags = append(tripTags, &ua.ReadValueID{NodeID: id})
		}
		tripReq := &ua.ReadRequest{
			NodesToRead:        tripTags,
			MaxAge:             2222,
			TimestampsToReturn: ua.TimestampsToReturnBoth,
		}

		// если есть файл с данными текущего часа, читаем его в структуру
		filedata, err := os.ReadFile(arhDirName + time.Now().Format("20060102_15") + ".json")
		if err == nil {
			err := json.Unmarshal(filedata, &d)
			if err != nil {
				log.Println(err)
			}
		}

		err = d.ReadOpcTagList(ctx, cl) // вычитываем параметры с единицами измерения, комментами согласно списку тэгов
		if err != nil {
			log.Println(err)
		}

		spin := false // тэг для фиксации факта наличия вращения

		rpmInd := 0 // ищем индекс тэга контроля оборотов
		for i := range d.Tag {
			if d.Tag[i].Name == "ST50_BZK" {
				rpmInd = i
			}
		}

		wg.Add(1)
		// рутина отвечает за запросы к серверу и складывание данных в файлы
		go func() {
			if b != nil {
				b.SendTxt("логер запущен")
			}
			defer wg.Done()
			ticker := time.NewTicker(time.Duration(cfg.StoreCycle) * time.Second) // тикер записи файлов
			chkSpin := time.NewTicker(6 * time.Second)                            // тикер для проверки оборотов
			current_hour := time.Now().Hour()

			for {
				select {
				case <-ctx.Done():
					log.Println("data process stopped")
					return
				case <-chkSpin.C:
					var curRpm float32
					if len(d.Tag[rpmInd].Y) > 0 {
						curRpm = d.Tag[rpmInd].Y[len(d.Tag[rpmInd].Y)-1].Value.(float32) // если нашли тэг оборотов, то зачитываем его
					}
					// момент запуска с очисткой данных
					if !spin && curRpm > 6.6 {
						spin = true
						d.Clean()
						if b != nil {
							b.SendTxt("запуськ пошоль")
						}
					}
					// момент останова с записью файла и отправкой через бота
					if spin && curRpm < 6.6 {
						spin = false
						buf, filename, err := repository.StoreData(&d, arhDirName, false)
						if err != nil {
							log.Println(err)
						}
						if b != nil {
							err := b.SendTxt("остановились")
							if err != nil {
								log.Println(err)
							}
							err = b.SendArh(buf, filename)
							if err != nil {
								log.Println(err)
							}
						}
					}

					if cl.State() == opcua.Connected && b != nil {
						resp, err := cl.Read(ctx, tripReq) //читаем тэги первопричин
						if err != nil {
							fmt.Println(err)
						}
						first := resp.Results[0].Value.Value().(uint32) // фиксируем вид останова

						if first != 0 && prevFirst == 0 {
							b.SendTxt(tripreport.GetFirst(resp)) // вычисляем первопричину и отправляем в телегу
						}
						prevFirst = first
					}

				case <-ticker.C:
					nowT := time.Now()

					if nowT.Hour() != current_hour && !spin {
						current_hour = nowT.Hour()
						// если пошол новый час, то пишем архив и чистим данные
						_, _, err := repository.StoreData(&d, arhDirName, true)
						if err != nil {
							log.Println(err)
						} else {
							d.Clean()
						}
					} else {
						// пишем данные в джисон-файл по тикеру
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
					// перебираем циклы и формируем обращения к серверу
					for key, item := range d.Ccs {
						// если пришло время обратиться, то обращаемся
						if item.Cct >= key {
							if cl.State() == opcua.Connected {
								item.Resp, err = cl.Read(ctx, item.Req)
							}
							if err != nil {
								log.Fatal("opcua request error: ", err)
							}

							item.Cct = 0
						}
						// заполняем слайсы новыми данными
						for i := range item.Resp.Results {
							crTm = item.Resp.Results[i].ServerTimestamp.Local().Format("15:04:05.000")
							d.AddV(item.FirstPos+i, item.Resp.Results[i].Value.Value().(float32), crTm)
						}

						item.Cct += d.MinCycle // для контроля момента обращения

						if item.Cct <= d.MinCycle {
							newTm = crTm
						}
					}
					if newTm != "" {
						d.Tm = append(d.Tm, newTm)
					}
					time.Sleep(time.Duration(d.MinCycle) * time.Millisecond) // ждем время минимального цикла
				}
			}
		}()
	}
	// хттп-сервер для отображения трендов
	mux := http.NewServeMux()
	srv := http.Server{
		Addr:    cfg.TrPort,
		Handler: mux,
	}
	stopSrvSig := make(chan struct{}) //сигнал остановки сервера

	wg.Add(1)
	// рутина для отслеживания сигнала остановки сервера
	go func() {
		defer wg.Done()
		<-stopSrvSig
		err := srv.Shutdown(ctx)
		if err != nil {
			log.Println(err)
		}
	}()

	wg.Add(1)
	// рутина с сервером
	go func() {
		defer wg.Done()

		mux.HandleFunc("/", trendView)
		err := srv.ListenAndServe()
		if err != nil {
			log.Println(err)
		}

	}()

	filename := ""
	// тут ждем файл с данными для просмотра
	for {
		if MdRd {
			fmt.Printf("для останова введи ку\nчто именно пялим > ")
		} else {
			fmt.Print("для останова введи ку > ")
		}
		fmt.Scan(&filename)
		if strings.TrimSpace(filename) == "q" { // или команду останова
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
		b.SendTxt("логер остановлен")
	}
	close(stopSrvSig) // отправляет сигнал останова хттп-серверу
	if cl != nil {
		cl.Close(ctx) // отключаем описи юа клиент
	}
	cancel() // отменяем контекст для завершения всех процессов

	wg.Wait() // ждем останова всех рутин
}

// хэндлер для отрисовки трендов
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
