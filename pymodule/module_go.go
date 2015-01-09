package pymodule

import (
	"fmt"

	"github.com/lsaint/py"

	"transcendence/network"
	"transcendence/proto"

	pb "code.google.com/p/goprotobuf/proto"
)

type GoModule struct {
	sendChan chan *proto.GateOutPack
	pm       *network.Postman
	isLeader bool
	tw       *TaskWaitress
}

func NewGoModule(out chan *proto.GateOutPack, pm *network.Postman, tw *TaskWaitress) *GoModule {
	mod := &GoModule{sendChan: out,
		pm:       pm,
		isLeader: false,
		tw:       tw}
	return mod
}

func (this *GoModule) Py_SendMsg(args *py.Tuple) (ret *py.Base, err error) {
	var sid, uri int
	var sbin string
	var lids []int
	err = py.ParseV(args, &sid, &uri, &sbin, &lids)
	if err != nil {
		fmt.Println("SendMsg err", err)
		return
	}

	lids64 := make([]int64, len(lids))
	for idx, i := range lids {
		lids64[idx] = int64(i)
	}

	//b, err := base64.StdEncoding.DecodeString(sbin)
	//if err != nil {
	//	fmt.Println("Base64 decode err", err)
	//	return
	//}
	this.sendChan <- &proto.GateOutPack{
		Lids: lids64,
		Sid:  pb.Uint32(uint32(sid)),
		Uri:  pb.Uint32(uint32(uri)),
		Bin:  []byte(sbin)}
	return py.IncNone(), nil
}

func (this *GoModule) Py_PostAsync(args *py.Tuple) (ret *py.Base, err error) {
	var url, content string
	var sn int
	err = py.Parse(args, &url, &content, &sn)
	this.pm.PostAsync(url, content, int64(sn))
	return py.IncNone(), nil
}

func (this *GoModule) Py_Post(args *py.Tuple) (ret *py.Base, err error) {
	var url, content string
	err = py.Parse(args, &url, &content)
	s := this.pm.Post(url, content)
	return py.NewString(s).Obj(), nil
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
