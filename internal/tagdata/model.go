package tagdata

import (
	"github.com/go-echarts/go-echarts/v2/opts"
)

type TagData struct {
	Name    string          `json:"name"`
	Dscr    string          `json:"dscr"`
	Min     float64         `json:"min"`
	Max     float64         `json:"max"`
	CycleMs int             `json:"cyclems"`
	Unit    string          `json:"senstype"`
	Y       []opts.LineData `json:"y"`
	T       []string        `json:"t"`
}
