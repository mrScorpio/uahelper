package tagdata

import (
	"github.com/go-echarts/go-echarts/v2/opts"
)

type TagData struct {
	Name     string          `json:"name"`
	Dscr     string          `json:"dscr"`
	Min      float32         `json:"min"`
	Max      float32         `json:"max"`
	CycleMs  int             `json:"cyclems"`
	SensType string          `json:"senstype"`
	Y        []opts.LineData `json:"y"`
	T        []string        `json:"t"`
}
