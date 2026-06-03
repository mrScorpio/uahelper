package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/unit"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/mrscorpio/uahelper/configs"
	"github.com/mrscorpio/uahelper/internal/repository"
	"github.com/mrscorpio/uahelper/internal/tagdata"
	"github.com/mrscorpio/uahelper/internal/trend"
	"github.com/mrscorpio/uahelper/internal/tripreport"
	"github.com/mrscorpio/uahelper/internal/ui"
	"github.com/mrscorpio/uahelper/pkg/opcuacl"
	"github.com/mrscorpio/uahelper/pkg/tgbot"
	"github.com/mrscorpio/uahelper/pkg/vkbot"
)

const MdRd bool = false // для выбора перед компиляцией - логер 0 или вьюер 1

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := http.NewServeMux()

	ui.NewData = make(chan string) //канал для передачи имени файла от уи в бэкенд
	ui.Cmd = make(chan int)
	ui.Gogo = true

	cfg := configs.LoadConfig()
	err := cfg.WrFile()
	if err != nil {
		log.Println(err)
	}

	arhDirName := "arh/" //папка для хранения файлов

	httpAddr := "http://localhost" + cfg.TrPort + "/?zoom=st50_bzk&show=zt504&step=1"

	b, err := tgbot.NewBot(ctx, cfg, MdRd) // бот, в режиме просмотра нил
	if err != nil {
		log.Println(err)
	}
	vkb, err := vkbot.NewBot(cfg, MdRd)
	if err != nil {
		log.Println(err)
	}

	cl, err := opcuacl.NewCl(ctx, cfg, MdRd) // описи юа клиент, в режиме просмотра нил
	if err != nil {
		log.Println(err)
	}

	d := new(tagdata.AllTags)
	var wTime time.Time
	legSel := make(map[string]bool) // для отключения позиций легенды в трендах

	var wg sync.WaitGroup

	if !MdRd {
		fmt.Println("читаем сервачок", cfg.Endpoint)
		os.Mkdir(strings.TrimSuffix(arhDirName, "/"), 0755)

		var prevFirst uint32 = 6 // чтобы не спамить аварией при запуске логера

		// тут тэги с первопричинами аварии
		tagname := []string{
			"PROT.FIRSTCOM.FIRST",
			"PROT.STARTDENY.FIRST",
			"PROT.F_UNLOAD.FIRST",
			"PROT.FORCE.FIRST",
			"PROT.TRIP.FIRST",
			"PROT.CCDENY.FIRST",
			"PROT.OILF.FIRST",
			"PROT.PURGEF.FIRST",
			"PROT.TRIPRES.FIRST",
			"PROT.TRIPMECH.FIRST",
			"PROT.TRIPTURB1.FIRST",
			"PROT.TRIPTURB2.FIRST",
			"PROT.TRIPETC.FIRST",
			"PROT.STARTDENYALG.FIRST",
		}
		tripTags := make([]*ua.ReadValueID, 0)
		for _, v := range tagname {
			id, err := ua.ParseNodeID("ns=2;s=Application." + v)
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
		fire := false // тэг для фиксации момента розжига

		rpmInd := 0 // ищем индекс тэга контроля оборотов
		for i := range d.Tag {
			if d.Tag[i].Name == "ST50_BZK" {
				rpmInd = i
			}
		}

		wg.Add(1)
		// рутина отвечает за запросы к серверу данных
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					log.Println("data process stopped")
					return

				default:
					newTm := ""
					crTm := ""
					// перебираем циклы и формируем обращения к серверу
					for key, item := range d.Ccs {
						// если пришло время обратиться, то обращаемся
						if item.Cct >= key {
							if cl[0].State() == opcua.Connected {
								clNum := 0
								if len(cl) > 1 {
									if cl[1] != nil {
										if cl[1].State() == opcua.Connected {
											clNum = 1 // если достучались до второго узла, то тянем данные с него
										}
									}
								}
								item.Resp, err = cl[clNum].Read(ctx, item.Req)
							}
							if err != nil {
								log.Println("opcua request error: ", err)
								continue
							}

							item.Cct = 0
						}
						// заполняем слайсы новыми данными
						for i := range item.Resp.Results {
							if i == len(item.Resp.Results)-1 {
								crTm = item.Resp.Results[0].ServerTimestamp.Local().Format("15:04:05.000")
							}
							v := item.Resp.Results[i].Value.Value()
							if v == nil {
								d.AddV(item.FirstPos+i, 6.6)
								fmt.Println("tag N", item.FirstPos+i, "has no data")
							} else {
								d.AddV(item.FirstPos+i, float64(v.(float32)))
							}
						}

						item.Cct += d.MinCycle // для контроля момента обращения

						if item.Cct <= d.MinCycle {
							newTm = crTm
						}
					}
					if newTm != "" {
						d.Mu.Lock()
						d.Tm = append(d.Tm, newTm)
						d.Mu.Unlock()
					}
					time.Sleep(time.Duration(d.MinCycle) * time.Millisecond) // ждем время минимального цикла
				}
			}
		}()

		wg.Add(1)
		// рутина отвечает складывание данных в файлы и отрисовку онлайн-тренда
		go func() {
			if b != nil {
				go b.SendToBoss("логер запущен")
			}
			defer wg.Done()
			ticker := time.NewTicker(time.Duration(cfg.StoreCycle) * time.Second) // тикер записи файлов
			chkSpin := time.NewTicker(3 * time.Second)                            // тикер для проверки оборотов

			currentHour := time.Now().Hour()

			for {
				select {
				case <-ctx.Done():
					log.Println("file process stopped")
					ui.Cmd <- 6
					return
				case <-chkSpin.C:
					var curRpm float64
					if len(d.Tag[rpmInd].Y) > 0 {
						curRpm = d.Tag[rpmInd].Y[len(d.Tag[rpmInd].Y)-1].Value.(float64) // если нашли тэг оборотов, то зачитываем его
					}
					// момент запуска с очисткой данных
					if !spin && curRpm > 666.666 {
						spin = true
						d.Clean()
						d.TripTM = time.Now().Add(time.Duration(66666) * time.Hour) // типа трип когда-то случится
						d.TripTag = "X3"
						mes := "раскрутка, смотреть на ingcgt.ru"
						if b != nil {
							b.SendTxt(mes)
						}
						if vkb != nil {
							err := vkb.B.NewTextMessage(cfg.VkChat, mes).Send()
							if err != nil {
								log.Println(err)
							}
						}
					}
					// момент розжига
					if !fire && curRpm > 5222.222 {
						fire = true
						mes := "есть розжиг"
						if b != nil {
							b.SendTxt(mes)
						}
						if vkb != nil {
							err := vkb.B.NewTextMessage(cfg.VkChat, mes).Send()
							if err != nil {
								log.Println(err)
							}
						}
					}
					// момент останова с записью файла и отправкой через бота
					if spin && curRpm < 6.6 {
						spin = false
						fire = false
						buf, filename, err := repository.StoreData(d, arhDirName, false)
						if err != nil {
							log.Println(err)
						}
						if b != nil {
							err := b.SendTxt("вращения нет")
							if err != nil {
								log.Println(err)
							}
							err = b.SendArh(buf, filename)
							if err != nil {
								log.Println(err)
							}
						}
						if vkb != nil {
							toSend, err := os.OpenFile(filename, os.O_RDONLY, 0755)
							if err != nil {
								log.Println(err)
							} else {
								err := vkb.B.NewFileMessage(cfg.VkChat, toSend).Send()
								if err != nil {
									log.Println(err)
								}
							}
						}
					}
					if len(cl) > 1 {
						if cl[1] != nil {
							if cl[1].State() == opcua.Connected {
								resp, err := cl[1].Read(ctx, tripReq) //читаем тэги первопричин
								if err != nil {
									fmt.Println(err)
									continue
								}
								first := resp.Results[0].Value.Value().(uint32) // фиксируем вид останова

								if first != 0 && prevFirst == 0 {
									mes, triptag := tripreport.GetFirst(resp) // вычисляем первопричину
									if b != nil {
										b.SendTxt(mes) // отправляем в телегу
									}
									if vkb != nil {
										err := vkb.B.NewTextMessage(cfg.VkChat, mes).Send()
										if err != nil {
											log.Println(err)
										}
									}
									d.TripTM = time.Now()
									d.TripTag = triptag
								}
								prevFirst = first
							}
						}
					}

				case <-ticker.C:
					nowT := time.Now()

					if nowT.Hour() != currentHour && !spin {
						currentHour = nowT.Hour()
						// если пошол новый час, то пишем архив и чистим данные
						_, _, err := repository.StoreData(d, arhDirName, true)
						if err != nil {
							log.Println(err)
						} else {
							d.Clean()
						}
						//check arch send
						/*
							if b != nil {
								err = b.SendArh(buf, filename)
								if err != nil {
									log.Println(err)
								}
							}
							if vkb != nil {
								toSend, err := os.OpenFile(arhDirName+filename, os.O_RDONLY, 0755)
								if err != nil {
									log.Println(err)
								} else {
									err := vkb.B.NewFileMessage(cfg.VkChat, toSend).Send()
									if err != nil {
										log.Println(err)
									}
								}
							}
						*/

					} else {
						// пишем данные в джисон-файл по тикеру
						data, err := json.Marshal(d)
						if err != nil {
							log.Println(err)
						} else {
							filename := arhDirName + nowT.Format("20060102_15") + ".json"
							err = os.WriteFile(filename, data, 0755)
							if err != nil {
								log.Println(err)
							}
						}

					}

				default:
					if ui.Gogo {
						err := ui.DrawChart(d, cfg)
						if err != nil {
							log.Println(err)
						}
					}
					time.Sleep(time.Duration(999) * time.Millisecond)

				}
			}
		}()
	}

	// хттп-сервер для отображения трендов
	srv := http.Server{
		Addr:    cfg.TrPort,
		Handler: mux,
	}
	stopSrvSig := make(chan struct{}) //сигнал остановки сервера

	wg.Add(1)
	// рутина для отслеживания сигнала остановки сервера
	go func() {
		defer wg.Done()
		mes := "логер остановлен"
		if vkb != nil {
			defer vkb.B.NewTextMessage(cfg.VkBossId, mes).Send()
		}
		if b != nil {
			defer b.SendToBoss(mes)
		}
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

		mux.Handle("/", trend.View(d, legSel, &wTime))
		err := srv.ListenAndServe()
		if err != nil {
			log.Println(err)
		}

	}()

	if MdRd {
		ui.BufImg = image.NewRGBA(image.Rect(0, 0, 22, 16))
		wg.Add(1)
		go func() {
			defer wg.Done()
			initDataLoad := true
			for name := range ui.NewData {
				wTime, err = repository.ReadStored(d, name)
				if initDataLoad {
					close(ui.DataLoaded)
					initDataLoad = false
				} else {
					ui.TL.UpdTaglist(d, cfg)
				}
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		w := new(app.Window)
		if MdRd {
			w.Option(app.Title("вьюер"))
		} else {
			w.Option(app.Title("логер"))
		}
		w.Option(app.Size(unit.Dp(600), unit.Dp(800)))
		if err := ui.DrawUi(w, d, cfg, MdRd); err != nil {
			log.Println(err)
		}
		log.Println("stop from ui")
		close(stopSrvSig) // отправляет сигнал останова хттп-серверу
		close(ui.NewData)
		if b != nil {
			time.Sleep(time.Duration(666) * time.Millisecond) // чтобы бот отправил сообщение об останове
		}
		for i := range cl {
			if cl[i] != nil {
				cl[i].Close(ctx) // отключаем описи юа клиент
			}
		}
		cancel() // отменяем контекст для завершения всех процессов
		wg.Done()
		wg.Wait() // ждем останова всех рутин
		os.Exit(0)
	}()

	go app.Main()

	conn, err := net.Dial("tcp", "ya.ru:80")
	myip := strings.Split(conn.LocalAddr().String(), ":")
	if err != nil {
		log.Println(err)
	} else {
		//httpAddr = "http://" + myip[0] + cfg.TrPort + "/?zoom=st50_bzk&show=zt504&step=1"
		httpAddr = ui.HttPath(d, cfg, myip[0])
	}
	conn.Close()

	fmt.Println("тренды пялить на", httpAddr)

	filename := ""
	// тут ждем файл с данными для просмотра
	for {
		if MdRd {
			cmd := exec.Command(cfg.BrPath, httpAddr)
			err := cmd.Start()
			if err != nil {
				log.Println(err)
			}
			fmt.Printf("что именно пялим > ")
		}
		fmt.Scan(&filename)
		if strings.TrimSpace(filename) == "q" { // или команду останова
			break
		}
		wTime, err = repository.ReadStored(d, filename)

		if err != nil {
			log.Println(err)
			continue
		}

		fmt.Println("загружено, смотри в браузере")
	}

	close(stopSrvSig) // отправляет сигнал останова хттп-серверу
	close(ui.NewData)

	for i := range cl {
		if cl[i] != nil {
			cl[i].Close(ctx) // отключаем описи юа клиент
		}
	}
	cancel() // отменяем контекст для завершения всех процессов

	wg.Wait() // ждем останова всех рутин
}
