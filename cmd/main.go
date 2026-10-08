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
	"github.com/mrscorpio/uahelper/internal/trend"
	"github.com/mrscorpio/uahelper/internal/tripreport"
	"github.com/mrscorpio/uahelper/internal/ui"
	"github.com/mrscorpio/uahelper/pkg/natscl"
	"github.com/mrscorpio/uahelper/pkg/opcuacl"
	"github.com/mrscorpio/uahelper/pkg/tagdata"
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
	//ui.Gogo = true

	cfg := configs.LoadConfig()

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

	dbCl, err := repository.DbConnect(cfg)
	if err != nil {
		log.Println("InfluxDB client driver:", err)
	}

	if dbCl != nil {
		chkInflux, err := dbCl.Cl.GetServerVersion()

		if err != nil {
			log.Println(err)
		} else {
			log.Println("Connected to InfluxDB ver.", chkInflux)
		}
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
			"GVL_PROTECT.F_FIRST.FIRST",
			"GVL_PROTECT.F_START.FIRST",
			"GVL_PROTECT.F_UNLOAD.FIRST",
			"GVL_PROTECT.F_FORCE.FIRST",
			"GVL_PROTECT.F_TRIP.FIRST",
			"GVL_PROTECT.F_CRANKING.FIRST",
			"GVL_PROTECT.F_OILCHECK.FIRST",
			"GVL_PROTECT.F_PURGE.FIRST",
			"GVL_PROTECT.F_RESERVE.FIRST",
			"GVL_PROTECT.F_TRIP_VIBR.FIRST",
			"GVL_PROTECT.F_TRIP_TURB.FIRST",
			"GVL_PROTECT.F_TRIP_BEARINGS.FIRST",
			"GVL_PROTECT.F_TRIP_ETC.FIRST",
			"GVL_PROTECT.F_START_SEQUENCE.FIRST",
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

		err = d.ReadOpcTagList(ctx, cl) // вычитываем параметры с единицами измерения, комментами согласно списку тэгов
		if err != nil {
			log.Println(err)
		}

		cleanData, err := d.CopyData(1)
		if err != nil {
			log.Println(err)
		}
		log.Println(len(cleanData.Tt), len(cleanData.Tag))

		// если есть файл с данными текущего часа, читаем его в структуру
		filedata, err := os.ReadFile(arhDirName + time.Now().Format("20060102_15") + ".json")
		if err == nil {
			err := json.Unmarshal(filedata, d)
			if err != nil {
				log.Println(err)
			}
		} else {
			d.Tag = append(d.Tag, d.NewTag("opcDly", "задержка ответа OPC", d.MinCycle))
			d.Tag[len(d.Tag)-1].Unit = "мс"
			d.Unit["мс"] = tagdata.NewUnit()
			d.Unit["мс"].Pos = append(d.Unit["мс"].Pos, len(d.Tag)-1)
		}

		cfg.UpdTagMap(d)
		err = cfg.WrFile()
		if err != nil {
			log.Println(err)
		}

		natsCl, err := natscl.NewNats(cfg.NatsAddr, d.GetTagNum())
		if err != nil {
			log.Println(err)
		}
		if natsCl != nil {
			err := natsCl.ListenCmd(d, cleanData)
			if err != nil {
				log.Println(err)
			}
		}

		spin := false // тэг для фиксации факта наличия вращения
		fire := false // тэг для фиксации момента розжига

		rpmInd := 0 // ищем индекс тэга контроля оборотов
		for i := range d.Tag {
			if d.Tag[i].Name == "ST50" {
				rpmInd = i
			}
		}

		wg.Add(1)
		// рутина отвечает за запросы к серверу данных
		go func() {
			defer wg.Done()
			crTm := time.Now()
			newTm := crTm
			sendNats := 0
			//chkReqCnt := 0
			var chkTm time.Time
			var crCycleMs, opcRespMs int64

			for {
				select {
				case <-ctx.Done():
					if natsCl != nil {
						natsCl.C.Close()
					}
					log.Println("data process stopped")

					return

				default:
					//var crTm time.Time
					// перебираем циклы и формируем обращения к серверу
					chkTm = time.Now()
					crCycleMs = 0
					nodes := make([]*ua.ReadValueID, 0)
					req := &ua.ReadRequest{
						NodesToRead:        nodes,
						MaxAge:             float64(cfg.OpcMaxAge),
						TimestampsToReturn: ua.TimestampsToReturnBoth,
					}
					resp := &ua.ReadResponse{}
					cclOrder := make([]int, 0)

					for key, item := range d.Ccs {
						// если пришло время обратиться, то добавляем тэги в запрос
						if item.Cct >= key {
							req.NodesToRead = append(req.NodesToRead, item.ReqTags...)
							cclOrder = append(cclOrder, key)
							item.Cct = 0
						}

						item.Cct += d.MinCycle // для контроля момента обращения
					}

					if len(req.NodesToRead) == 0 {
						continue
					}

					if cl[0].State() == opcua.Connected {
						clNum := 0
						if len(cl) > 1 {
							if cl[1] != nil {
								if cl[1].State() == opcua.Connected {
									clNum = 1 // если достучались до второго узла, то тянем данные с него
								}
							}
						}
						reqctx, reqCancel := context.WithTimeout(ctx, time.Duration(cfg.OpcCtxTmout)*time.Millisecond)
						resp, err = cl[clNum].Read(reqctx, req)
						reqCancel()
						opcRespMs = time.Since(chkTm).Milliseconds()
						/*
							if resp != nil && cfg.RdMd {
								fmt.Print(reqDur, "ms-n=", len(resp.Results), " ")
							}
						*/
						if opcRespMs > 666 {
							log.Println("ua response is fucking slow -", opcRespMs, "ms to answer")
						}
						if err != nil {
							log.Println("opcua request error: ", err)
							d.AddV(len(d.Tag)-1, float32(opcRespMs))
							continue
						}
					}
					resLen := len(resp.Results)
					if resLen > 0 {
						newTm = resp.Results[0].ServerTimestamp.Local()

						for i := range d.Tag {
							d.AddPreV(i)
						}

						i := 0
						for _, cc := range cclOrder {
							for j := 0; j < d.Ccs[cc].Q; j++ {
								v := resp.Results[i].Value.Value()
								if v != nil {
									ind := d.Ccs[cc].FirstPos + j
									d.ChgLastV(ind, v.(float32))
									if dbCl != nil {
										dbCl.Data[d.Tag[ind].Name] = v.(float32)
									}
									if natsCl != nil {
										natsCl.OnlineBuf[ind] = v.(float32)
									}
								}
								i++
							}
						}
						//добавили значение задержки опроса
						d.ChgLastV(len(d.Tag)-1, float32(opcRespMs))
						if dbCl != nil {
							lasTag := len(d.Tag) - 1
							dbCl.Data[d.Tag[lasTag].Name] = opcRespMs // и в базу
						}
					}
					if natsCl != nil {
						natsCl.TimeBuf = newTm
					}

					if dbCl != nil {
						if err := dbCl.DbWr(ctx, newTm); err != nil {
							log.Println(err)
						}
					}

					if d.AddT(newTm, spin) && !ui.Gogo && !MdRd {
						if ui.LastInd > 666 {
							ui.LastInd -= 666
						}
						if ui.FstInd > 666 {
							ui.FstInd -= 666
						}
					}

					if natsCl != nil && sendNats > 9 {
						sendNats = 0
						err := natsCl.SendCurrent()
						if err != nil {
							log.Println(err)
						}
					}
					sendNats++
					crCycleMs = time.Since(chkTm).Milliseconds()
					/*if chkReqCnt > 222 {
						fmt.Println(" =", crCycleMs)
						chkReqCnt = 0
					}
					chkReqCnt++
					*/
					if crCycleMs > 666 {
						log.Println("cycle is fucking slow -", crCycleMs, "ms")
					}
					if crCycleMs < int64(d.MinCycle) {
						//fmt.Println(" wait", d.MinCycle-int(crCycleMs))
						time.Sleep(time.Duration(d.MinCycle-int(crCycleMs)) * time.Millisecond) // ждем время минимального цикла
					}
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
					err := repository.SaveJson(d, arhDirName, time.Now())
					if err != nil {
						log.Println(err)
					}
					/*
						_, _, err = repository.StoreData(d, arhDirName, true)
						if err != nil {
							log.Println(err)
						}
					*/
					if dbCl != nil {
						err = dbCl.Cl.Close()
						if err != nil {
							log.Println(err)
						} else {
							log.Println("InfluxDb connection closed")
						}
					}
					log.Println("file process stopped")
					ui.Cmd <- 6
					return
				case <-chkSpin.C:
					var curRpm float32
					if len(d.Tag[rpmInd].V) > 0 {
						curRpm = d.Tag[rpmInd].V[len(d.Tag[rpmInd].V)-1] // если нашли тэг оборотов, то зачитываем его
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
							err := vkb.B.NewTextMessage(cfg.VkChat, "вращения нет").Send()
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
						spin = false
						fire = false
					}
					if len(cl) > 1 {
						if cl[1] != nil {
							if cl[1].State() == opcua.Connected {
								resp, err := cl[1].Read(ctx, tripReq) //читаем тэги первопричин
								if err != nil {
									fmt.Println(err)
									continue
								}
								f := resp.Results[0].Value.Value() // фиксируем вид останова
								if f == nil {
									continue
								}
								first := f.(uint32)

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
						} /*else {	!больше не чистим!
							d.Clean()
						}*/
						ui.CrearHoursData()

					} else {
						// пишем данные в джисон-файл по тикеру
						err := repository.SaveJson(d, arhDirName, nowT)
						if err != nil {
							log.Println(err)
						}
						/*
							d.Mu.RLock()
							data, err := json.Marshal(d)
							d.Mu.RUnlock()
							if err != nil {
								log.Println(err)
							} else {
								filename := arhDirName + nowT.Format("20060102_15") + ".json"
								err = os.WriteFile(filename, data, 0755)
								if err != nil {
									log.Println(err)
								}
							}
						*/
					}

				default:
					if ui.Gogo {
						err := ui.DrawChart(cfg)
						if err != nil {
							log.Println(err)
						}
					}

					time.Sleep(time.Duration(100) * time.Millisecond)
				}
			}
		}()
	}

	//if MdRd {
	ui.BufImg = image.NewRGBA(image.Rect(0, 0, 22, 16))
	wg.Add(1)
	go func() {
		defer wg.Done()
		initDataLoad := true
		for name := range ui.NewData {
			if MdRd {
				wTime, err = repository.ReadStored(d, name)
				if initDataLoad {
					close(ui.DataLoaded)
					initDataLoad = false
				} else {
					ui.TL.UpdTaglist(d, cfg)
					ui.DrawChart(cfg)
				}
			} else {
				ui.Mu.Lock()
				ui.Gogo = false
				*ui.PrevHour = 666
				if ui.TrendData[*ui.PrevHour] == nil {
					ui.TrendData[*ui.PrevHour] = new(tagdata.AllTags)
				}
				wTime, err = repository.ReadStored(ui.TrendData[*ui.PrevHour], name)
				ui.Mu.Unlock()
				ui.DrawChart(cfg)
			}
		}
	}()
	//}

	// хттп-сервер для отображения трендов
	srv := http.Server{
		Addr:    cfg.TrPort,
		Handler: mux,
	}
	stopSrvSig := make(chan struct{}) //сигнал остановки сервера

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

		mux.Handle("/echart/", trend.View(ui.TrendData, ui.PrevHour, legSel, &wTime))
		err := srv.ListenAndServe()
		if err != nil {
			log.Println(err)
		}

	}()

	if err != nil {
		log.Println(err)
	} else {
		//httpAddr = "http://" + myip[0] + cfg.TrPort + "/?zoom=st50_bzk&show=zt504&step=1"
		myip := strings.Split(conn.LocalAddr().String(), ":")
		httpAddr = ui.HttPath(d, cfg, myip[0])
		conn.Close()
	}

	fmt.Println("тренды пялить на", httpAddr)

	filename := ""
	// тут ждем файл с данными для просмотра
	for {
		if MdRd {
			cmd := exec.Command(cfg.BrPath, httpAddr)
			if filename != "" {
				err := cmd.Start()
				if err != nil {
					log.Println(err)
				}
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
