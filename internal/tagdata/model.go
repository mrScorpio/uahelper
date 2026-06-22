package tagdata

type TagData struct {
	Name    string    `json:"name"`
	Dscr    string    `json:"dscr"`
	Min     float64   `json:"min"`
	Max     float64   `json:"max"`
	CycleMs int       `json:"cyclems"`
	Unit    string    `json:"senstype"`
	V       []float64 `json:"v"`
	//Y       []opts.LineData `json:"-"`
	//T       []string        `json:"t"`
}
