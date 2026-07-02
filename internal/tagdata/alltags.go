package tagdata

import (
	"bufio"
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

type UnitData struct {
	Pos []int
	Max float64
	Min float64
}

func NewUnit() *UnitData {
	pos := make([]int, 0)
	return &UnitData{
		Pos: pos,
	}
}

type AllTags struct {
	Tag      []*TagData
	Unit     map[string]*UnitData
	Descr    map[string]string
	Ccs      map[int]*CycleData
	MinCycle int
	Tt       []time.Time
	TripTM   time.Time
	TripTag  string
	Mu       sync.RWMutex `json:"-"`
}

func (at *AllTags) NewTag(name string, dscr string, cycle int) *TagData {
	if at.Tag[0] == nil {
		at.Descr = make(map[string]string)
		//at.Tm = make([]string, 0, 6666)
		at.Tt = make([]time.Time, 0, 6666)
	}
	//t := make([]string, 0, 6)
	//y := make([]opts.LineData, 0, 6666)
	v := make([]float64, 0, 6666)
	at.Descr[name] = dscr
	at.Mu.Lock()
	defer at.Mu.Unlock()
	return &TagData{
		Name: name,
		Dscr: dscr,
		//Y:    y,
		//T:       t,
		V:       v,
		CycleMs: cycle,
	}
}

func (at *AllTags) AddV(i int, v float64) {
	if v > 66666.66666 {
		v = 0.0
	}
	if v < -66666.66666 {
		v = 0.0
	}
	at.Mu.Lock()
	at.Tag[i].V = append(at.Tag[i].V, v)
	defer at.Mu.Unlock()
	//	at.Tag[i].T = append(at.Tag[i].T, t)
	unit := at.Tag[i].Unit
	if len(at.Tag[i].V) == 1 {
		at.Tag[i].Max = v
	}

	if at.Tag[i].Min == 0.0 {
		at.Tag[i].Min = at.Tag[i].Max

		if at.Unit[unit].Min == 0.0 {
			at.Unit[unit].Min = at.Unit[unit].Max
		}
	}

	if v > at.Tag[i].Max {
		at.Tag[i].Max = v
		if at.Tag[i].Max > at.Unit[unit].Max {
			at.Unit[unit].Max = at.Tag[i].Max
		}
	}

	if v < at.Tag[i].Min {
		at.Tag[i].Min = v
		if at.Tag[i].Min < at.Unit[unit].Min {
			at.Unit[unit].Min = at.Tag[i].Min
		}
	}

}

func (at *AllTags) Clean() {
	at.Mu.Lock()
	for i := range at.Tag {
		//at.Tag[i].V = make([]float64, 0, 6666)
		if len(at.Tag[i].V) > 22222 {
			at.Tag[i].V = at.Tag[i].V[len(at.Tag[i].V)-22222:]
		}
	}
	//at.Tt = make([]time.Time, 0, 6666)
	if len(at.Tt) > 22222 {
		at.Tt = at.Tt[len(at.Tt)-22222:]
	}
	at.Mu.Unlock()
}

func (at *AllTags) AddT(newT time.Time, w bool) bool {
	cut := false
	at.Mu.Lock()
	defer at.Mu.Unlock()
	at.Tt = append(at.Tt, newT)
	if newT.Sub(at.Tt[0]) > time.Duration(66*time.Minute) && !w {
		for i := range at.Tag {
			if len(at.Tag[i].V) > 666 {
				at.Tag[i].V = at.Tag[i].V[666:]
			}
		}
		if len(at.Tt) > 666 {
			at.Tt = at.Tt[666:]
			cut = true
		}
	}
	return cut
}

func (d *AllTags) ReadOpcTagList(ctx context.Context, cl []*opcua.Client) error {
	if cl[0] == nil {
		return nil
	}

	tagname := []string{}
	cycle := []int{}

	tagfile, err := os.Open("tags")
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(tagfile)

	d.Ccs = make(map[int]*CycleData)
	nextCycle := 222
	d.MinCycle = 666
	maxCycle := 6
	i := 0

	preTag := "ns=1;s=REGUL_R500."
	postTag := ".VALUE"

	if len(cl) > 1 {
		if cl[1] != nil {
			if cl[1].State() == opcua.Connected {
				preTag = "ns=2;s=Application."
				postTag = ".OUT.VALUE"
			}
		}
	}

	for scanner.Scan() {
		c, err := strconv.Atoi(strings.TrimSuffix(scanner.Text(), ":"))
		if err != nil {
			nextTag := scanner.Text()
			tagname = append(tagname, nextTag)
			cycle = append(cycle, nextCycle)

			err := d.Ccs[nextCycle].AddTag(preTag + nextTag + postTag)
			if err != nil {
				return err
			}
			i++
		} else {
			nextCycle = c
			d.Ccs[nextCycle] = NewCycle()
			d.Ccs[nextCycle].FirstPos = i
			if c < d.MinCycle {
				d.MinCycle = c
			}
			if c > maxCycle {
				maxCycle = c
			}
		}
	}
	tagfile.Close()

	id := make([]*ua.NodeID, len(tagname))
	uid := make([]*ua.NodeID, len(tagname))
	node := make([]*opcua.Node, len(tagname))

	unitsToRead := make([]ua.ReadValueID, len(tagname))
	unitsToReadp := make([]*ua.ReadValueID, len(tagname))

	newTags := false
	if len(d.Tag) != len(tagname) {
		d.Tag = make([]*TagData, len(tagname))
		newTags = true
	}

	for i, v := range tagname {

		id[i], err = ua.ParseNodeID("ns=1;s=REGUL_R500." + v + ".VALUE")
		if err != nil {
			log.Fatalf("invalid node id: %v", err)
			return err
		}

		uid[i], err = ua.ParseNodeID("ns=1;s=REGUL_R500." + v + ".EU")
		if err != nil {
			log.Fatalf("invalid node id: %v", err)
			return err
		}

		if newTags {
			node[i] = cl[0].Node(id[i])
			descr, err := node[i].Description(ctx)
			if err != nil {
				log.Fatal(err)
				return err
			}

			fullTag := strings.Split(v, ".")
			if len(fullTag) > 1 {
				d.Tag[i] = d.NewTag(fullTag[1], descr.Text, cycle[i])
			} else {
				d.Tag[i] = d.NewTag(v, descr.Text, cycle[i])
			}
		}

		unitsToRead[i].NodeID = uid[i]
		unitsToReadp[i] = &unitsToRead[i]

	}

	reqUnits := &ua.ReadRequest{
		MaxAge:      2000,
		NodesToRead: unitsToReadp,
	}

	var resp *ua.ReadResponse

	if newTags {
		resp, err = cl[0].Read(ctx, reqUnits)
		if err != nil {
			log.Fatal(err)
		}

		d.Unit = make(map[string]*UnitData)

		for i, v := range resp.Results {
			key := v.Value.Value().(string)
			if key == "°С" {
				key = "°C"
			}
			d.Tag[i].Unit = key
			_, ok := d.Unit[key]
			if !ok {
				d.Unit[key] = NewUnit()
			}
			d.Unit[key].Pos = append(d.Unit[key].Pos, i)
		}
	}

	return nil
}

func (at *AllTags) ChangeId() error {
	for j := range at.Ccs {
		firstPos := at.Ccs[j].FirstPos
		for i := range at.Ccs[j].ReqTags {
			ids := "ns=2;s=Application.AI." + at.Tag[firstPos+i].Name + ".OUT.VALUE"
			id, err := ua.ParseNodeID(ids)
			if err != nil {
				return err
			}
			at.Ccs[j].ReqTags[i] = &ua.ReadValueID{NodeID: id}
		}
	}
	return nil
}
