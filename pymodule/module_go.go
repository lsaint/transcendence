package pymodule

import (
	"fmt"

	"github.com/qiniu/py"

	"transcendence/network"
	"transcendence/proto"

	pb "code.google.com/p/goprotobuf/proto"
)

type GoModule struct {
	sendChan chan *proto.Passpack
	pm       *network.Postman
	isLeader bool
	tw       *TaskWaitress
}

func NewGoModule(out chan *proto.Passpack, pm *network.Postman, tw *TaskWaitress) *GoModule {
	mod := &GoModule{sendChan: out,
		pm:       pm,
		isLeader: false,
		tw:       tw}
	return mod
}

func (this *GoModule) Py_SendMsg(args *py.Tuple) (ret *py.Base, err error) {
	var tsid, ssid, uri, action, fid int
	var sbin string
	var uids []int
	err = py.ParseV(args, &tsid, &ssid, &uri, &sbin, &action, &fid, &uids)
	if err != nil {
		fmt.Println("SendMsg err", err)
		return
	}

	uids32 := make([]uint32, len(uids))
	for idx, i := range uids {
		uids32[idx] = uint32(i)
	}

	//b, err := base64.StdEncoding.DecodeString(sbin)
	//if err != nil {
	//	fmt.Println("Base64 decode err", err)
	//	return
	//}
	this.sendChan <- &proto.Passpack{Tsid: pb.Uint32(uint32(tsid)),
		Ssid:   pb.Uint32(uint32(ssid)),
		Uri:    pb.Uint32(uint32(uri)),
		Action: proto.Action(action).Enum(),
		Uids:   uids32,
		Fid:    pb.Uint32(uint32(fid)),
		Bin:    []byte(sbin)}
	return py.IncNone(), nil
}

func (this *GoModule) Py_PostAsync(args *py.Tuple) (ret *py.Base, err error) {
	var url, content string
	var sn int
	err = py.Parse(args, &url, &content, &sn)
	this.pm.PostAsync(url, content, int64(sn))
	return py.IncNone(), nil
}

func (this *GoModule) Py_IsLeader(args *py.Tuple) (ret *py.Base, err error) {
	if this.isLeader {
		return py.NewInt(1).Obj(), nil
	} else {
		return py.NewInt(0).Obj(), nil
	}
}

func (this *GoModule) Py_PyReady(args *py.Tuple) (*py.Base, error) {
	this.tw.PyReady()
	return py.IncNone(), nil
}

func (this *GoModule) Py_DoTask(args *py.Tuple) (*py.Base, error) {
	this.tw.DoTask()
	return py.IncNone(), nil
}
