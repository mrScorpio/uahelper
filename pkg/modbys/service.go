package modbys

import (
	"time"

	"github.com/simonvetter/modbus"
)

func NewCl(url string) (*modbus.ModbusClient, error) {
	cl, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     url,
		Timeout: 666 * time.Millisecond,
	})
	if err != nil {
		return nil, err
	}
	err = cl.Open()
	if err != nil {
		return nil, err
	}
	return cl, nil
}
