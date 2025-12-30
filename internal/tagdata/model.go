package tagdata

import (
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/gopcua/opcua/ua"
)

type CycleData struct {
	Q        int
	ReqTags  []*ua.ReadValueID
	FirstPos int
	Req      *ua.ReadRequest
	Resp     *ua.ReadResponse
	Cct      int
}

type AllTags struct {
	Tag   []*TagData
	Unit  map[string]*UnitData
	Descr map[string]string
}

type UnitData struct {
	Pos []int
	Max float32
	Min float32
}

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

func (at *AllTags) NewTag(name string, dscr string, cycle int) *TagData {
	if at.Tag[0] == nil {
		at.Descr = make(map[string]string)
	}
	t := make([]string, 0, 6666)
	y := make([]opts.LineData, 0, 6666)
	at.Descr[name] = dscr
	return &TagData{
		Name:    name,
		Dscr:    dscr,
		Y:       y,
		T:       t,
		CycleMs: cycle,
	}
}

func NewUnit() *UnitData {
	pos := make([]int, 0)
	return &UnitData{
		Pos: pos,
	}
}

func NewCycle() *CycleData {
	req := make([]*ua.ReadValueID, 0)
	return &CycleData{
		ReqTags: req,
		Req: &ua.ReadRequest{
			NodesToRead:        req,
			MaxAge:             2222,
			TimestampsToReturn: ua.TimestampsToReturnBoth,
		},
		Resp: &ua.ReadResponse{},
	}
}

func (cd *CycleData) AddTag(tagname string) error {
	id, err := ua.ParseNodeID("ns=1;s=REGUL_R500." + tagname + ".VALUE")
	if err != nil {
		return err
	}
	cd.Q++
	cd.ReqTags = append(cd.ReqTags, &ua.ReadValueID{NodeID: id})
	cd.Req.NodesToRead = cd.ReqTags
	return nil
}

func (at *AllTags) AddV(i int, v float32, t string) {
	at.Tag[i].Y = append(at.Tag[i].Y, opts.LineData{Value: v})
	at.Tag[i].T = append(at.Tag[i].T, t)
	unit := ""
	if len(at.Tag[i].Y) == 1 {
		at.Tag[i].Max = v
	}

	if v > at.Tag[i].Max || v < at.Tag[i].Min {
	l1:
		for key := range at.Unit {
			for _, v := range at.Unit[key].Pos {
				if i == v {
					unit = key
					break l1
				}
			}
		}

		if v > at.Tag[i].Max {
			at.Tag[i].Max = v
			if at.Tag[i].Max > at.Unit[unit].Max {
				at.Unit[unit].Max = at.Tag[i].Max
			}
		}

		if at.Tag[i].Min == 0.0 {
			at.Tag[i].Min = at.Tag[i].Max
			if at.Unit[unit].Min == 0.0 {
				at.Unit[unit].Min = at.Unit[unit].Max
			}
		}

		if v < at.Tag[i].Min {
			at.Tag[i].Min = v
			if at.Tag[i].Min < at.Unit[unit].Min {
				at.Unit[unit].Min = at.Tag[i].Min
			}
		}
	}
}

func (at *AllTags) Clean() {
	for i := range at.Tag {
		at.Tag[i].T = make([]string, 0, 6666)
		at.Tag[i].Y = make([]opts.LineData, 0, 6666)
	}
}
