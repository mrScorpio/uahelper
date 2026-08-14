package natscl

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mrscorpio/uahelper/pkg/tagdata"
	"github.com/nats-io/nats.go"
)

type NatsCl struct {
	C         *nats.Conn
	OnlineBuf []float32
	ClCmd     []byte
	Clients   map[string]byte
	ClNum     byte
}

func NewNats(addr string, tagNum int) (*NatsCl, error) {
	cl, err := nats.Connect(addr)
	if err != nil {
		return nil, err
	}
	buf := make([]float32, tagNum)
	clients := make(map[string]byte)
	return &NatsCl{C: cl, OnlineBuf: buf, Clients: clients}, nil
}

func (nc *NatsCl) SendCurrent() error {
	var buf bytes.Buffer
	sendBuf := make([]int32, len(nc.OnlineBuf))
	for i, v := range nc.OnlineBuf {
		sendBuf[i] = int32(v * 1000)
	}
	err := binary.Write(&buf, binary.BigEndian, sendBuf)
	if err != nil {
		return err
	}
	err = nc.C.Publish("online", buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (nc *NatsCl) ListenCmd(d *tagdata.AllTags) error {
	_, err := nc.C.Subscribe("cmd", func(msg *nats.Msg) {
		var clId string
		var err error

		nc.ClCmd = msg.Data

		for i := 1; i < len(nc.ClCmd); i++ {
			clId += fmt.Sprint(nc.ClCmd[i])
		}
		_, ok := nc.Clients[clId]
		if !ok {
			nc.ClNum++
			nc.Clients[clId] = nc.ClNum
		}

		if nc.ClCmd[0] == 66 {
			err = nc.InitNewClient(clId, d)
			if err != nil {
				log.Println(err)
			}
		}
		log.Println("command ", nc.ClCmd[0], "receved from client", nc.Clients[clId])
	})
	if err != nil {
		return err
	}
	return nil
}

func (nc *NatsCl) InitNewClient(clId string, d *tagdata.AllTags) error {

	d.Mu.RLock()
	data, err := json.MarshalIndent(d, "", "  ")
	d.Mu.RUnlock()
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, err := w.Create("initdata.json")
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	w.Close()

	err = nc.C.Publish(clId, buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}
