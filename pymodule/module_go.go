package pymodule

import (
	"encoding/base64"
	"fmt"

	"github.com/qiniu/py"

	"transcendence/network"
	"transcendence/proto"

	pb "code.google.com/p/goprotobuf/proto"
)

type GoModule struct {
	sendChan chan *proto.Passpack
	pm       *network.Postman
}

func NewGoModule(out chan *proto.Passpack, pm *network.Postman) *GoModule {
	mod := &GoModule{sendChan: out, pm: pm}
	return mod
}

func (this *GoModule) Py_SendMsg(args *py.Tuple) (ret *py.Base, err error) {
	var tsid, ssid, uri, action, fid int
	var sbin string
	var uids []int
	var uids32 []uint32
	err = py.ParseV(args, &tsid, &ssid, &uri, &sbin, &action, &fid, &uids)
	if err != nil {
		fmt.Println("SendMsg err", err)
		return
	}

	for idx, i := range uids {
		uids32[idx] = uint32(i)
	}

	b, err := base64.StdEncoding.DecodeString(sbin)
	if err != nil {
		fmt.Println("Base64 decode err", err)
		return
	}
	this.sendChan <- &proto.Passpack{Tsid: pb.Uint32(uint32(tsid)),
		Ssid:   pb.Uint32(uint32(ssid)),
		Uri:    pb.Uint32(uint32(uri)),
		Action: proto.Action(action).Enum(),
		Uids:   uids32,
		Fid:    pb.Uint32(uint32(fid)),
		Bin:    b}
	return py.IncNone(), nil
}

func (this *GoModule) Py_PostAsync(args *py.Tuple) (ret *py.Base, err error) {
	var url, content string
	var sn int
	err = py.Parse(args, &url, &content, &sn)
	this.pm.PostAsync(url, content, int64(sn))
	return py.IncNone(), nil
}
