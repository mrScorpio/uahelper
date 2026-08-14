package tagdata

type TagData struct {
	Name    string    `json:"name"`
	Dscr    string    `json:"dscr"`
	Min     float32   `json:"min"`
	Max     float32   `json:"max"`
	CycleMs int       `json:"cyclems"`
	Unit    string    `json:"senstype"`
	V       []float32 `json:"v"`
	//Y       []opts.LineData `json:"-"`
	//T       []string        `json:"t"`
}
