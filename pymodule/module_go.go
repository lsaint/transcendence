package pymodule

import (
	"encoding/base64"
	"fmt"

	"github.com/qiniu/py"

	"transcendence/postman"
	"transcendence/proto"

	pb "code.google.com/p/goprotobuf/proto"
)

type GoModule struct {
	sendChan chan *proto.GateOutPack
	pm       *postman.Postman
}

func NewGoModule(out chan *proto.GateOutPack, pm *postman.Postman) *GoModule {
	mod := &GoModule{sendChan: out, pm: pm}
	return mod
}

func (this *GoModule) Py_SendMsg(args *py.Tuple) (ret *py.Base, err error) {
	var tsid, ssid, uri, action, uid, fid int
	var sbin string
	err = py.Parse(args, &tsid, &ssid, &uri, &sbin, &action, &uid, &fid)
	if err != nil {
		fmt.Println("SendMsg err", err)
		return
	}

	b, err := base64.StdEncoding.DecodeString(sbin)
	if err != nil {
		fmt.Println("Base64 decode err", err)
		return
	}
	this.sendChan <- &proto.GateOutPack{Tsid: pb.Uint32(uint32(tsid)),
		Ssid:   pb.Uint32(uint32(ssid)),
		Uri:    pb.Uint32(uint32(uri)),
		Action: proto.Action(action).Enum(),
		Uid:    pb.Uint32(uint32(uid)),
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
