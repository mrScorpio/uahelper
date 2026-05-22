package tripreport

import (
	"math"

	"github.com/gopcua/opcua/ua"
)

func GetFirst(resp *ua.ReadResponse) (string, string) {
	res := "почему-то встали, надо разбираться)"
	tagname := "X3"
	if len(resp.Results) > 0 {
		switch resp.Results[0].Value.Value().(uint32) {
		case 1:
			res = "Отмена пуска\n"
		case 2:
			res = "Снятие нагрузки"
		case 4:
			res = "Вынужденный останов"
		case 8:
			res = "Аварийный останов"
			if len(resp.Results) > 4 {
				switch resp.Results[4].Value.Value().(uint32) {
				case 1:
					if len(resp.Results) > 9 {
						triptype := resp.Results[9].Value.Value().(uint32)
						switch triptype {
						case uint32(math.Pow(2, 0)):
							tagname = "VT507"
						case uint32(math.Pow(2, 1)):
							tagname = "ZT503"
						case uint32(math.Pow(2, 2)):
							tagname = "VT508"
						case uint32(math.Pow(2, 3)):
							tagname = "ZT504"
						case uint32(math.Pow(2, 4)):
							tagname = "VT509"
						case uint32(math.Pow(2, 5)):
							tagname = "VT505"
						case uint32(math.Pow(2, 6)):
							tagname = "ZT501"
						case uint32(math.Pow(2, 7)):
							tagname = "VT506"
						case uint32(math.Pow(2, 8)):
							tagname = "ZT502"
						case uint32(math.Pow(2, 9)):
							tagname = "VT510"
						case uint32(math.Pow(2, 10)):
							tagname = "VT501"
						case uint32(math.Pow(2, 11)):
							tagname = "VT502"
						case uint32(math.Pow(2, 12)):
							tagname = "VT503"
						case uint32(math.Pow(2, 13)):
							tagname = "VT504"
						case uint32(math.Pow(2, 14)):
							tagname = "TT576"
						case uint32(math.Pow(2, 15)):
							tagname = "TT577"
						case uint32(math.Pow(2, 16)):
							tagname = "TT578"
						case uint32(math.Pow(2, 17)):
							tagname = "TT579"
						case uint32(math.Pow(2, 18)):
							tagname = "TT580"
						case uint32(math.Pow(2, 19)):
							tagname = "TT581"
						case uint32(math.Pow(2, 20)):
							tagname = "TT582"
						case uint32(math.Pow(2, 21)):
							tagname = "TT583"
						case uint32(math.Pow(2, 22)):
							tagname = "TT590"
						case uint32(math.Pow(2, 23)):
							tagname = "TT591"
						case uint32(math.Pow(2, 24)):
							tagname = "TT592"
						case uint32(math.Pow(2, 25)):
							tagname = "TT593"
						case uint32(math.Pow(2, 26)):
							tagname = "TT594"
						case uint32(math.Pow(2, 27)):
							tagname = "TT595"
						case uint32(math.Pow(2, 28)):
							tagname = "TT596"
						case uint32(math.Pow(2, 29)):
							tagname = "TT597"
						}
						if triptype < 16384 {
							res += "\nвибрация жесть по датчику " + tagname
						} else {
							res += "\nтемпература жесть по датчику " + tagname
						}
					}
				case 2:
					res += " по параметрам турбины\n"
					if len(resp.Results) > 8 {
						triptype := resp.Results[8].Value.Value().(uint32)
						switch triptype {
						case uint32(math.Pow(2, 0)):
							tagname = "PDT720"
							res += "высокое давление газа на выхлопе"
						case uint32(math.Pow(2, 1)):
							tagname = "PDT721"
							res += "высокое давление газа на выхлопе"
						case uint32(math.Pow(2, 2)):
							tagname = "PDT722"
							res += "высокое давление газа на выхлопе"
						case uint32(math.Pow(2, 3)):
							tagname = "PDT723"
							res += "высокое давление газа на выхлопе"
						case uint32(math.Pow(2, 4)):
							tagname = "PDT902"
							res += "высокое разрежение воздуха на всасе"
						case uint32(math.Pow(2, 5)):
							tagname = "TE501A"
							res += "высокая температура колодок опорного подшипника"
						case uint32(math.Pow(2, 6)):
							tagname = "TE501B"
							res += "высокая температура колодок опорного подшипника"
						case uint32(math.Pow(2, 7)):
							tagname = "TE502A"
							res += "высокая температура колодок упорного подшипника"
						case uint32(math.Pow(2, 8)):
							tagname = "TE502B"
							res += "высокая температура колодок упорного подшипника"
						case uint32(math.Pow(2, 9)):
							tagname = "TE503A"
							res += "высокая температура колодок упорного подшипника"
						case uint32(math.Pow(2, 10)):
							tagname = "TE503B"
							res += "высокая температура колодок упорного подшипника"
						case uint32(math.Pow(2, 11)):
							tagname = "TE504A"
							res += "высокая температура колодок упорного подшипника"
						case uint32(math.Pow(2, 12)):
							tagname = "TE504B"
							res += "высокая температура колодок упорного подшипника"
						case uint32(math.Pow(2, 13)):
							tagname = "TE717"
							res += "высокая температура на выхлопе"
						case uint32(math.Pow(2, 14)):
							tagname = "TE511"
							res += "проскок пламени в камере сгорания 1"
						case uint32(math.Pow(2, 15)):
							tagname = "TE512"
							res += "проскок пламени в камере сгорания 2"
						case uint32(math.Pow(2, 16)):
							tagname = "TE513"
							res += "проскок пламени в камере сгорания 3"
						case uint32(math.Pow(2, 17)):
							tagname = "TE514"
							res += "проскок пламени в камере сгорания 4"
						case uint32(math.Pow(2, 18)):
							tagname = "TE515"
							res += "проскок пламени в камере сгорания 5"
						case uint32(math.Pow(2, 19)):
							tagname = "TT581"
						case uint32(math.Pow(2, 20)):
							tagname = "TT582"
						case uint32(math.Pow(2, 21)):
							tagname = "TT583"
						case uint32(math.Pow(2, 22)):
							tagname = "TT590"
						case uint32(math.Pow(2, 23)):
							tagname = "TT591"
						case uint32(math.Pow(2, 24)):
							tagname = "TT592"
						case uint32(math.Pow(2, 25)):
							tagname = "TT593"
						case uint32(math.Pow(2, 26)):
							tagname = "TT594"
						case uint32(math.Pow(2, 27)):
							tagname = "TT595"
						case uint32(math.Pow(2, 28)):
							tagname = "TT596"
						case uint32(math.Pow(2, 29)):
							tagname = "TT597"
						}
						res += " - датчик " + tagname
					}
				case 4:
					res += "\nсработал блок защиты турбины"
				case 8:
					if len(resp.Results) > 12 {
						switch resp.Results[12].Value.Value().(uint32) {
						case 1:
							res += "\nотказ системы пожарной автоматики"
						case 2:
							res += "\nаварийная загазованность"
						case 4:
							res += "\nнет напряжения на вводе ИБП"
						case 8:
							res += "\nавария выпрямителя ИБП"
						case 16:
							res += "\nавария байпасной линии ИБП"
						case 32:
							res += "\nавария генератора"
						case 64:
							res += "\nавария в блоке обеспечения"
						case 128:
							res += "\nсработка высоковольтной релейной защиты"
						case 256:
							res += "\nаварийная кнопка на пульте резервного управления"
						case 512:
							res += "\nот системы вибромониторинга ТИК"
						case 1024:
							res += "\nпревышение допустимой скорости вращения"
						case 2048:
							res += "\nаварийная кнопка на шкафу управления"
						case 4096:
							res += "\nавария дозатора толива"
						case 8192:
							res += "\nнизкий воздухообмен в системе обдува привода"
						case 16384:
							res += "\nотказ системы контроля загазованности"
						case 32768:
							res += "\nнет реакции от выключателя СН генератора"
						case 65536:
							res += "\nнет реакции от выключателя НО генератора"
						case 131072:
							res += "\nнет реакции от выключателя СНа генератора"
						case 262144:
							res += "\nнет реакции от выключателя ВО генератора"
						case 524288:
							res += "\nПожар"
						}
					}
				}
			}
		case 16:
			res = "Отмена ХП"
		case 32:
			res = "Экстренный останов"
		case 64:
			res = "Ошибка прогрева масла"
		case 128:
			res = "Ошибка продувки КП"
		case 256:
			res = "АО в резерве"
		}
	}
	return res, tagname
}
