package tagdata

import (
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
