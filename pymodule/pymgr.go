package pymodule

import (
	"encoding/base64"
	"log"
	"time"

	"transcendence/conf"
	"transcendence/network"
	"transcendence/postman"
	"transcendence/proto"

	"github.com/qiniu/py"
)

const (
	PROTO_INVOKE  = iota
	UPDATE_INVOKE = iota
	NET_CTRL      = iota
	POST_DONE     = iota
)

type PyMgr struct {
	recvChan chan *proto.GateInPack
	sendChan chan *proto.GateOutPack
	httpChan chan *network.HttpReq
	pm       *postman.Postman
	glue     *py.Module

	gomod    py.GoModule
	logmod   py.GoModule
	redismod py.GoModule
	salmod   py.GoModule
}

func NewPyMgr(in chan *proto.GateInPack, out chan *proto.GateOutPack,
	http_req_chan chan *network.HttpReq) *PyMgr {

	mgr := &PyMgr{recvChan: in,
		httpChan: http_req_chan,
		sendChan: out,
		pm:       postman.NewPostman()}
	var err error
	mgr.gomod, err = py.NewGoModule("go", "", NewGoModule(out, mgr.pm))
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

	code, err := py.CompileFile("./script/glue.py", py.FileInput)
	if err != nil {
		log.Fatalln("Compile failed:", err)
	}
	defer code.Decref()

	mgr.glue, err = py.ExecCodeModule("glue", code.Obj())
	if err != nil {
		log.Fatalln("ExecCodeModule glue err:", err)
	}

	mgr.salmod, err = py.NewGoModule("sal", "", NewSalModule(conf.CF.SVCTYPE, mgr.glue))
	if err != nil {
		log.Fatalln("ExecCodeModule glue err:", err)
	}

	_, err = mgr.glue.CallMethodObjArgs("test")
	// defer mgr.pymode.Decref()
	if err != nil {
		log.Fatalln("ExecCodeModule failed:", err)
	}
	return mgr
}

func (this *PyMgr) Start() {
	ticker := time.Tick(1 * time.Second)
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
		}
	}
}

func (this *PyMgr) onProto(pack *proto.GateInPack) {
	tsid := py.NewInt64(int64(pack.GetTsid()))
	defer tsid.Decref()
	ssid := py.NewInt64(int64(pack.GetSsid()))
	defer ssid.Decref()
	uri := py.NewInt(int(pack.GetUri()))
	defer uri.Decref()
	uid := py.NewInt64(int64(pack.GetUid()))
	defer uid.Decref()
	b := base64.StdEncoding.EncodeToString(pack.Bin)
	data := py.NewString(string(b))
	defer data.Decref()
	_, err := this.glue.CallMethodObjArgs("OnProto", tsid.Obj(), ssid.Obj(),
		uri.Obj(), data.Obj(), uid.Obj())
	if err != nil {
		log.Println("OnProto err:", err)
	}
}

func (this *PyMgr) onTicker() {
	if _, err := this.glue.CallMethodObjArgs("OnTicker"); err != nil {
		log.Println("onTicker err:", err)
	}
}

func (this *PyMgr) onPostDone(sn int64, ret string) {
	py_sn := py.NewInt64(sn)
	defer py_sn.Decref()
	py_ret := py.NewString(string(ret))
	defer py_ret.Decref()
	if _, err := this.glue.CallMethodObjArgs("OnPostDone", py_sn.Obj(), py_ret.Obj()); err != nil {
		log.Println("onPostDone err:", err)
	}
}

func (this *PyMgr) onHttpReq(jn, url string) string {
	py_jn := py.NewString(jn)
	defer py_jn.Decref()
	py_url := py.NewString(url)
	defer py_url.Decref()
	r, err := this.glue.CallMethodObjArgs("OnHttpReq", py_jn.Obj(), py_url.Obj())

	if err != nil {
		log.Println("onHttpReq err:", err)
		return ""
	}
	if ret, ok := py.AsString(r); ok {
		return ret.String()
	}
	return ""
}
