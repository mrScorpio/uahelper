package tagdata

import (
	"bufio"
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

type UnitData struct {
	Pos []int
	Max float32
	Min float32
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
	Tm       []string
}

func (at *AllTags) NewTag(name string, dscr string, cycle int) *TagData {
	if at.Tag[0] == nil {
		at.Descr = make(map[string]string)
		at.Tm = make([]string, 0, 6666)
	}
	t := make([]string, 0, 6)
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

func (at *AllTags) AddV(i int, v float32, t string) {
	if v > 66666.66666 {
		v = 0.0
	}
	if v < -66666.66666 {
		v = 0.0
	}
	at.Tag[i].Y = append(at.Tag[i].Y, opts.LineData{Value: v})
	//	at.Tag[i].T = append(at.Tag[i].T, t)
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
	at.Tm = make([]string, 0, 6666)
}

func (d *AllTags) ReadOpcTagList(ctx context.Context, cl *opcua.Client) error {
	if cl == nil {
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
	for scanner.Scan() {
		c, err := strconv.Atoi(strings.TrimSuffix(scanner.Text(), ":"))
		if err != nil {
			nextTag := scanner.Text()
			tagname = append(tagname, nextTag)
			cycle = append(cycle, nextCycle)

			err := d.Ccs[nextCycle].AddTag(nextTag)
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
		}

		uid[i], err = ua.ParseNodeID("ns=1;s=REGUL_R500." + v + ".EU")
		if err != nil {
			log.Fatalf("invalid node id: %v", err)
		}

		if newTags {
			node[i] = cl.Node(id[i])
			descr, err := node[i].Description(ctx)
			if err != nil {
				log.Fatal(err)
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
		resp, err = cl.Read(ctx, reqUnits)
		if err != nil {
			log.Fatal(err)
		}

		d.Unit = make(map[string]*UnitData)

		for i, v := range resp.Results {
			key := v.Value.Value().(string)
			if key == "°С" {
				key = "°C"
			}
			_, ok := d.Unit[key]
			if !ok {
				d.Unit[key] = NewUnit()
			}
			d.Unit[key].Pos = append(d.Unit[key].Pos, i)
		}
	}

	return nil
}
