package pymodule

import (
	"encoding/base64"
	"log"
	"time"

	"transcendence/network"
	"transcendence/proto"

	"github.com/hashicorp/raft"
	"github.com/lsaint/py"
)

const (
	PROTO_INVOKE  = iota
	UPDATE_INVOKE = iota
	NET_CTRL      = iota
	POST_DONE     = iota
)

type PyMgr struct {
	recvChan chan *proto.GateInPack
	httpChan chan *network.HttpReq

	pm *network.Postman
	cn *network.ClusterNode

	tw   *TaskWaitress
	glue *py.Module

	gomo       *GoModule
	gomod      py.GoModule
	logmod     py.GoModule
	redismod   py.GoModule
	salmod     py.GoModule
	raftmod    py.GoModule
	servicemod py.GoModule

	waiting chan bool
	pyready chan bool
}

func NewPyMgr(in chan *proto.GateInPack,
	out chan *proto.GateOutPack,
	http_req_chan chan *network.HttpReq) *PyMgr {

	waiting, pyready := make(chan bool), make(chan bool)
	mgr := &PyMgr{recvChan: in,
		httpChan: http_req_chan,
		cn:       network.NewClusterNode(),
		tw:       NewTaskWaitress(waiting, pyready),
		waiting:  waiting,
		pyready:  pyready,
		pm:       network.NewPostman()}
	var err error
	mgr.gomo = NewGoModule(out, mgr.pm, mgr.tw)
	mgr.gomod, err = py.NewGoModule("go", "", mgr.gomo)
	if err != nil {
		log.Fatalln("NewGoModule failed:", err)
	}

	mgr.logmod, err = py.NewGoModule("log", "", NewLogModule())
	if err != nil {
		log.Fatalln("NewLogModule failed:", err)
	}

	mgr.redismod, err = py.NewGoModule("redigo", "", NewRedisModule())
	if err != nil {
		log.Fatalln("NewRedisModule failed:", err)
	}

	//mgr.salmod, err = py.NewGoModule("sal", "", NewSalModule(conf.CF.SVCTYPE))
	//if err != nil {
	//	log.Fatalln("ExecCodeModule sal err:", err)
	//}

	mgr.raftmod, err = py.NewGoModule("raft", "", NewRaftModule(mgr.cn))
	if err != nil {
		log.Fatalln("NewRaftModule failed:", err)
	}

	code, err := py.CompileFile("./script/glue.py", py.FileInput)
	if err != nil {
		log.Fatalln("Compile failed:", err)
	}
	defer code.Decref()

	mgr.servicemod, err = py.NewGoModule("service", "", NewServiceModule(mgr))
	if err != nil {
		log.Fatalln("NewServiceModule failed:", err)
	}

	mgr.glue, err = py.ExecCodeModule("glue", code.Obj())
	if err != nil {
		log.Fatalln("ExecCodeModule glue err:", err)
	}

	_, err = mgr.glue.CallMethodObjArgs("test_script")
	if err != nil {
		log.Fatalln("ExecCodeModule failed:", err)
	}

	go func() {
		fd := py.NewInt64(mgr.tw.GetFd())
		fd.Decref()
		mgr.glue.CallMethodObjArgs("main", fd.Obj())
	}()

	return mgr
}

func (this *PyMgr) Start() {
	ticker := time.Tick(1 * time.Second)
	<-this.pyready
	for {
		select {
		case <-ticker:
			this.onTicker()
		case pack := <-this.recvChan:
			this.onProto(pack)
		case post_ret := <-this.pm.DoneChan:
			this.onPostDone(post_ret.Sn, <-post_ret.Ret)
		case req := <-this.httpChan:
			req.Ret <- this.onHttpReq(req.Req, req.Url)
		case ev := <-this.cn.NodeEventChan:
			this.onClusterNodeEvent(ev)
		case rlog := <-this.cn.RaftAgent.ApplyCh:
			this.onRaftApply(rlog)
		}
	}
}

func (this *PyMgr) onProto(pack *proto.GateInPack) {
	sid := py.NewInt64(int64(pack.GetSid()))
	defer sid.Decref()
	uri := py.NewInt(int(pack.GetUri()))
	defer uri.Decref()

	b := base64.StdEncoding.EncodeToString(pack.Bin)
	data := py.NewString(string(b))
	defer data.Decref()
	_, err := this.callPyFunc("OnGateProto", sid.Obj(), uri.Obj(), data.Obj())
	if err != nil {
		log.Println("OnGateProto err:", err)
	}
}

func (this *PyMgr) onTicker() {
	if _, err := this.callPyFunc("OnTicker"); err != nil {
		log.Println("onTicker err:", err)
	}
}

func (this *PyMgr) onPostDone(sn int64, ret string) {
	py_sn := py.NewInt64(sn)
	defer py_sn.Decref()
	py_ret := py.NewString(string(ret))
	defer py_ret.Decref()
	if _, err := this.callPyFunc("OnPostDone", py_sn.Obj(), py_ret.Obj()); err != nil {
		log.Println("onPostDone err:", err)
	}
}

func (this *PyMgr) onHttpReq(jn, url string) string {
	py_jn := py.NewString(jn)
	defer py_jn.Decref()
	py_url := py.NewString(url)
	defer py_url.Decref()
	r, err := this.callPyFunc("OnHttpReq", py_jn.Obj(), py_url.Obj())

	if err != nil {
		log.Println("onHttpReq err:", err)
		return ""
	}
	if ret, ok := py.AsString(r); ok {
		return ret.String()
	}
	return ""
}

func (this *PyMgr) onClusterNodeEvent(ev network.NodeEvent) {
	if ev.Event == network.NodeBecomeLeader {
		this.gomo.isLeader = true
	} else if ev.Event == network.NodeHandoffLeader {
		this.gomo.isLeader = false
	}

	py_ev_type := py.NewInt64(int64(ev.Event))
	defer py_ev_type.Decref()
	name := ""
	if ev.Node != nil {
		name = ev.Node.Name
	}
	py_node_name := py.NewString(name)
	defer py_node_name.Decref()

	_, err := this.callPyFunc("OnClusterNodeEvent", py_ev_type.Obj(), py_node_name.Obj())
	if err != nil {
		log.Println("OnClusterNodeEvent err:", err)
	}
}

func (this *PyMgr) callPyFunc(name string, args ...*py.Base) (*py.Base, error) {
	this.tw.Notify()
	<-this.waiting
	defer func() {
		this.waiting <- true
	}()
	return this.glue.CallMethodObjArgs(name, args...)
}

func (this *PyMgr) onRaftApply(rlog *raft.Log) {
	py_data := py.NewString(string(rlog.Data))
	defer py_data.Decref()

	_, err := this.callPyFunc("OnRaftApply", py_data.Obj())
	if err != nil {
		log.Println("OnRaftApply err:", err)
	}
}
