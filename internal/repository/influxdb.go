package repository

import (
	"context"
	"time"

	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
	"github.com/influxdata/line-protocol/v2/lineprotocol"
	"github.com/mrscorpio/uahelper/configs"
)

type InfluxCl struct {
	Cl     *influxdb3.Client
	Data   map[string]any
	points []*influxdb3.Point
}

func DbConnect(cfg *configs.Config) (*InfluxCl, error) {
	cl, err := influxdb3.New(influxdb3.ClientConfig{
		Host: cfg.InfluxUrl, Token: cfg.InfluxToken, Database: "stend",
	})
	if err != nil {
		return nil, err
	}
	data := make(map[string]any)
	points := make([]*influxdb3.Point, 0, 666)
	return &InfluxCl{Cl: cl, Data: data, points: points}, nil
}

func (ic *InfluxCl) DbWr(ctx context.Context, tm time.Time) error {
	point := influxdb3.NewPoint("gtdata",
		map[string]string{"sys": "gtd"},
		ic.Data,
		tm,
	)
	ic.points = append(ic.points, point)
	if len(ic.points) > 660 {
		sendPoints := make([]*influxdb3.Point, len(ic.points))
		copy(sendPoints, ic.points)
		if ic.Cl != nil {
			go ic.Cl.WritePoints(ctx, sendPoints, influxdb3.WithPrecision(lineprotocol.Millisecond))
		}
		ic.points = make([]*influxdb3.Point, 0, 666)
	}
	//fmt.Println(ic.Data)
	return nil
}
