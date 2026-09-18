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
		case uint32(math.Pow(2, 1)):
			res = "Отмена пуска"
			if len(resp.Results) > 1 {
				switch resp.Results[1].Value.Value().(uint32) {
				case uint32(math.Pow(2, 2)):
					res += " по алгоритму\n"
					if len(resp.Results) > 13 {
						triptype := resp.Results[13].Value.Value().(uint32)
						switch triptype {
						case uint32(math.Pow(2, 0)):
							res += "нет питания на стартере"
						case uint32(math.Pow(2, 1)):
							res += "стартер не крутит"
						case uint32(math.Pow(2, 2)):
							res += "неисправность отсечных клапанов"
						case uint32(math.Pow(2, 3)):
							res += "неисправен клапан топливной линии XY401"
						case uint32(math.Pow(2, 4)):
							res += "неисправен клапан топливной линии XY402"
						case uint32(math.Pow(2, 5)):
							res += "неисправен клапан топливной линии XY403"
						case uint32(math.Pow(2, 6)):
							res += "нет готовности блока обеспечения"
						case uint32(math.Pow(2, 7)):
							res += "не запустился блок обеспечения"
						case uint32(math.Pow(2, 8)):
							res += "не включился аппарат зажигания"
						case uint32(math.Pow(2, 9)):
							res += "ТК не вышел на ХХ"
						}
					}
				case uint32(math.Pow(2, 0)):
					res += " - выхлоп"
				case uint32(math.Pow(2, 1)):
					res += " - проблемы с горением"
				case uint32(math.Pow(2, 3)):
					res += " от блока защиты турбины"
				case uint32(math.Pow(2, 4)):
					res += " аварийное отключение стартера"
				}
			}
		case uint32(math.Pow(2, 7)):
			res = "Снятие нагрузки"
		case uint32(math.Pow(2, 3)):
			res = "Вынужденный останов"
		case uint32(math.Pow(2, 5)):
			res = "Аварийный останов"
			if len(resp.Results) > 4 {
				switch resp.Results[4].Value.Value().(uint32) {
				case uint32(math.Pow(2, 0)):
					if len(resp.Results) > 9 {
						triptype := resp.Results[9].Value.Value().(uint32)
						switch triptype {
						case uint32(math.Pow(2, 0)):
							tagname = "VT507"
						case uint32(math.Pow(2, 1)):
							tagname = "VT508"
						case uint32(math.Pow(2, 2)):
							tagname = "VT509"
						case uint32(math.Pow(2, 3)):
							tagname = "ZT503"
						case uint32(math.Pow(2, 4)):
							tagname = "ZT504"
						case uint32(math.Pow(2, 5)):
							tagname = "VT505"
						case uint32(math.Pow(2, 6)):
							tagname = "VT506"
						case uint32(math.Pow(2, 7)):
							tagname = "VT510"
						case uint32(math.Pow(2, 8)):
							tagname = "ZT501"
						case uint32(math.Pow(2, 9)):
							tagname = "ZT502"
						case uint32(math.Pow(2, 10)):
							tagname = "VT501"
						case uint32(math.Pow(2, 11)):
							tagname = "VT502"
						case uint32(math.Pow(2, 12)):
							tagname = "VT503"
						case uint32(math.Pow(2, 13)):
							tagname = "VT504"
						case uint32(math.Pow(2, 14)):
							tagname = "VT511"
						case uint32(math.Pow(2, 15)):
							tagname = "VT512"
						case uint32(math.Pow(2, 16)):
							tagname = "VT513"
						case uint32(math.Pow(2, 17)):
							res += "\nсигнал АВАРИЯ от системы вибромониторинга ТИК"
						case uint32(math.Pow(2, 18)):
							res += "\nОВЕРСПИД"
						}
						if triptype < uint32(math.Pow(2, 17)) {
							res += "\nвибрация жесть по датчику " + tagname
						}
					}
				case uint32(math.Pow(2, 3)):
					res += " по параметрам турбины\n"
					if len(resp.Results) > 10 {
						triptype := resp.Results[10].Value.Value().(uint32)
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
							tagname = "TE717"
							res += "высокая температура на выхлопе"
						case uint32(math.Pow(2, 6)):
							tagname = "TE511"
							res += "проскок пламени в камере сгорания 1"
						case uint32(math.Pow(2, 7)):
							tagname = "TE512"
							res += "проскок пламени в камере сгорания 2"
						case uint32(math.Pow(2, 8)):
							tagname = "TE513"
							res += "проскок пламени в камере сгорания 3"
						case uint32(math.Pow(2, 9)):
							tagname = "TE514"
							res += "проскок пламени в камере сгорания 4"
						case uint32(math.Pow(2, 10)):
							tagname = "TE515"
							res += "проскок пламени в камере сгорания 5"
						/*
							case uint32(math.Pow(2, 19)):
								tagname = "TE521"
								res += "погасло пламя в камере сгорания 1"
							case uint32(math.Pow(2, 20)):
								tagname = "TE531"
								res += "погасло пламя в камере сгорания 1"
							case uint32(math.Pow(2, 21)):
								tagname = "TE522"
								res += "погасло пламя в камере сгорания 2"
							case uint32(math.Pow(2, 22)):
								tagname = "TE532"
								res += "погасло пламя в камере сгорания 2"
							case uint32(math.Pow(2, 23)):
								tagname = "TE523"
								res += "погасло пламя в камере сгорания 3"
							case uint32(math.Pow(2, 24)):
								tagname = "TE533"
								res += "погасло пламя в камере сгорания 3"
							case uint32(math.Pow(2, 25)):
								tagname = "TE524"
								res += "погасло пламя в камере сгорания 4"
							case uint32(math.Pow(2, 26)):
								tagname = "TE534"
								res += "погасло пламя в камере сгорания 4"
							case uint32(math.Pow(2, 27)):
								tagname = "TE525"
								res += "погасло пламя в камере сгорания 5"
							case uint32(math.Pow(2, 28)):
								tagname = "TE535"
								res += "погасло пламя в камере сгорания 5"
						*/
						case uint32(math.Pow(2, 11)):
							tagname = "TE584"
							res += "горячо на выхлопе камеры сгорания 4"
						case uint32(math.Pow(2, 12)):
							tagname = "TE585"
							res += "горячо на выхлопе камеры сгорания 5"
						case uint32(math.Pow(2, 13)):
							tagname = "Gair"
							res += "низкий расход воздуха"
						case uint32(math.Pow(2, 14)):
							tagname = "PT701"
							res += "низкое давление воздуха за компрессором"
						case uint32(math.Pow(2, 15)):
							tagname = "PT701"
							res += "высокое давление воздуха за компрессором"
						case uint32(math.Pow(2, 16)):
							tagname = "TE701"
							res += "высокая температура воздуха за компрессором"
						case uint32(math.Pow(2, 17)):
							tagname = "PT406"
							res += "низкое давление диффузионного газа"
						case uint32(math.Pow(2, 18)):
							tagname = "PT406"
							res += "высокое давление диффузионного газа"
						case uint32(math.Pow(2, 19)):
							tagname = "PT407"
							res += "низкое давление топлива на предв. смешивание"
						case uint32(math.Pow(2, 20)):
							tagname = "PT407"
							res += "высокое давление топлива на предв. смешивание"
						}
						res += " - датчик " + tagname
					}
				case uint32(math.Pow(2, 3)):
					res += " по параметрам подшипников\n"
					if len(resp.Results) > 11 {
						triptype := resp.Results[11].Value.Value().(uint32)
						switch triptype {
						case uint32(math.Pow(2, 0)):
							tagname = "TE501A"
							res += "высокая температура колодок опорного подшипника"
						case uint32(math.Pow(2, 1)):
							tagname = "TE501B"
							res += "высокая температура колодок опорного подшипника"
						case uint32(math.Pow(2, 2)):
							tagname = "TE502A"
							res += "высокая температура колодок упорного подшипника"
						case uint32(math.Pow(2, 3)):
							tagname = "TE502B"
							res += "высокая температура колодок упорного подшипника"
						case uint32(math.Pow(2, 4)):
							tagname = "TE503A"
							res += "высокая температура колодок упорного подшипника"
						case uint32(math.Pow(2, 5)):
							tagname = "TE503B"
							res += "высокая температура колодок упорного подшипника"
						case uint32(math.Pow(2, 6)):
							tagname = "TE504A"
							res += "высокая температура колодок упорного подшипника"
						case uint32(math.Pow(2, 7)):
							tagname = "TE504B"
							res += "высокая температура колодок упорного подшипника"
						case uint32(math.Pow(2, 8)):
							tagname = "TT576"
							res += "высокая температура подшипника редуктора"
						case uint32(math.Pow(2, 9)):
							tagname = "TT577"
							res += "высокая температура подшипника редуктора"
						case uint32(math.Pow(2, 10)):
							tagname = "TT578"
							res += "высокая температура подшипника редуктора"
						case uint32(math.Pow(2, 11)):
							tagname = "TT579"
							res += "высокая температура подшипника редуктора"
						case uint32(math.Pow(2, 12)):
							tagname = "TT580"
							res += "высокая температура подшипника редуктора"
						case uint32(math.Pow(2, 13)):
							tagname = "TT581"
							res += "высокая температура подшипника редуктора"
						case uint32(math.Pow(2, 14)):
							tagname = "TT582"
							res += "высокая температура подшипника редуктора"
						case uint32(math.Pow(2, 15)):
							tagname = "TT583"
							res += "высокая температура подшипника редуктора"
						/*
							case uint32(math.Pow(2, 0)):
								tagname = "TT590"
							case uint32(math.Pow(2, 1)):
								tagname = "TT591"
							case uint32(math.Pow(2, 2)):
								tagname = "TT592"
							case uint32(math.Pow(2, 3)):
								tagname = "TT593"
							case uint32(math.Pow(2, 4)):
								tagname = "TT594"
							case uint32(math.Pow(2, 5)):
								tagname = "TT595"
						*/
						case uint32(math.Pow(2, 16)):
							tagname = "TT596"
							res += "высокая температура приводного подшипника генератора"
						case uint32(math.Pow(2, 17)):
							tagname = "TT597"
							res += "высокая температура неприводного подшипника генератора"
						}
						res += " - датчик " + tagname
					}
				case uint32(math.Pow(2, 5)):
					res += " от блока защиты турбины"
				case uint32(math.Pow(2, 4)):
					if len(resp.Results) > 12 {
						switch resp.Results[12].Value.Value().(uint32) {
						case uint32(math.Pow(2, 0)):
							res += "\nотказ системы пожарной автоматики"
						case uint32(math.Pow(2, 1)):
							res += "\nаварийная загазованность"
						case uint32(math.Pow(2, 2)):
							res += "\nнет напряжения на вводе ИБП"
						case uint32(math.Pow(2, 3)):
							res += "\nавария выпрямителя ИБП"
						case uint32(math.Pow(2, 4)):
							res += "\nавария байпасной линии ИБП"
						case uint32(math.Pow(2, 5)):
							res += "\nавария генератора"
						case uint32(math.Pow(2, 6)):
							res += "\nавария в блоке обеспечения"
						case uint32(math.Pow(2, 7)):
							res += "\nсработка высоковольтной релейной защиты"
						case uint32(math.Pow(2, 8)):
							res += "\nаварийная кнопка на пульте резервного управления"
						case uint32(math.Pow(2, 9)):
							res += "\nаварийная кнопка на шкафу управления"
						case uint32(math.Pow(2, 10)):
							res += "\nавария дозатора толива"
						case uint32(math.Pow(2, 11)):
							res += "\nнизкий воздухообмен в системе обдува привода"
						case uint32(math.Pow(2, 12)):
							res += "\nотказ системы контроля загазованности"
						case uint32(math.Pow(2, 13)):
							res += "\nнет реакции от выключателя СН генератора"
						case uint32(math.Pow(2, 14)):
							res += "\nнет реакции от выключателя НО генератора"
						case uint32(math.Pow(2, 15)):
							res += "\nнет реакции от выключателя СНа генератора"
						case uint32(math.Pow(2, 16)):
							res += "\nнет реакции от выключателя ВО генератора"
						}
					}
				case uint32(math.Pow(2, 6)):
					res += "\nПожар"
				}
			}
		case uint32(math.Pow(2, 10)):
			res = "Отмена ХП"
		case uint32(math.Pow(2, 9)):
			res = "Экстренный останов"
		case uint32(math.Pow(2, 8)):
			res = "Ошибка прогрева масла"
		case uint32(math.Pow(2, 2)):
			res = "Ошибка продувки КП"
		case uint32(math.Pow(2, 0)):
			res = "Нажата кнопка АО в режиме резерва"
		}
	}
	return res, tagname
}
