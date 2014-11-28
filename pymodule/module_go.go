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
	taskmgr	*TaskMgr
}

func NewGoModule(out chan *proto.Passpack, pm *network.Postman, taskmgr *TaskMgr) *GoModule {
	mod := &GoModule{sendChan: out, pm: pm, isLeader: false, taskmgr: taskmgr}
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

func (this *GoModule) Py_GetTask(args *py.Tuple) (*py.Base, error) {
	lt := this.taskmgr.GetTask()
	all := py.NewTuple(len(lt))
	//defer all.Decref()
	for i, task := range lt {
		item := py.NewTuple(1 + len(task.Args))
		item.SetItem(0, task.Name)
		//defer task.Name.Decref()
		if task.Args != nil {
			for j, arg := range task.Args {
				arg.Decref()
				item.SetItem(j+1, arg)
			}
		}
		all.SetItem(i, item.Obj())
		//defer item.Decref()
	}
	//fmt.Println("all", all)
	return all.Obj(), nil
}
