package tripreport

import "github.com/gopcua/opcua/ua"

func GetFirst(resp *ua.ReadResponse) string {
	res := "почему-то встали, надо разбираться)"
	if len(resp.Results) > 0 {
		switch resp.Results[0].Value.Value().(uint32) {
		case 1:
			res = "Отмена пуска\n"
		case 2:
			res = "Снятие нагрузки"
		case 4:
			res = "Вынужденный останов"
		case 8:
			res = "Аварийный останов\n"
			if len(resp.Results) > 1 {
				switch resp.Results[1].Value.Value().(uint32) {
				case 1:
					res += "по параметрам ПВД и подшипников"
				case 2:
					res += "по параметрам турбины"
				case 4:
					res += "БЗК"
				case 8:
					res += "вибрация жесть!"
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
	return res
}
