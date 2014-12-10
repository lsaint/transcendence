package pymodule

import (
	"encoding/json"
	"log"
	"transcendence/network"

	"github.com/qiniu/py"

	. "transcendence/conf"
)

type PyFuncCaller interface {
	callPyFunc(string, ...*py.Base) (*py.Base, error)
}

type ServiceModule struct {
	pm      *network.Postman
	httpSrv *network.HttpServer
	caller  PyFuncCaller
}

func NewServiceModule(caller PyFuncCaller) *ServiceModule {
	recvChan := make(chan *network.HttpReq, CF.BUF_QUEUE)
	service := &ServiceModule{
		pm:      network.NewPostman(),
		caller:  caller,
		httpSrv: network.NewHttpServer(recvChan, []string{"/uplinkmsg", "/leaveplatform"}),
	}
	service.init()
	return service
}

func (this *ServiceModule) init() {
	this.register()
}

/*
CTX_REG = {"appid":"120","appname":"testreghttp","port":"20000","groupid":"4500","isp":"3",
	"isp2ip":[{"ip":"121.14.170.11","isp":"2"}], "regkey":"1c4814e155bd8e0acdcc3d5748e952de"}
*/
func (this *ServiceModule) register() {
	ret := this.pm.Post(CF.URL_SERVICE_REG, CF.CTX_REG)
	if ret == "" {
		log.Fatalln("register to service timeout")
	}
	log.Println("[SERVICE]register sucess", ret)
}

func (this *ServiceModule) processServiceMsg() {
	for req := range this.httpSrv.ReqChan {
		switch req.Url {
		case "/uplinkmsg":
			this.uplinkmsg(req.Req, req.Ret)
		case "/leaveplatform":
			this.leaveplatform(req.Req, req.Ret)
		}
	}
}

type S_uplinkmsg struct {
	Msgfrom      string `json:"msgfrom"`
	Action       string `json:"action"`
	Appid        int64  `json:"appid"`
	Suid         int64  `json:"suid"`
	Uid          int64  `json:"uid"`
	Topsid       int64  `json:"topsid"`
	Subsid       int64  `json:"subsid"`
	Yyfrom       string `json:"yyfrom"`
	Terminaltype int64  `json:"terminaltype"`
	Userip       string `json:"userip"`
	Data         string `json:"data"`
}

func (this *ServiceModule) uplinkmsg(req string, reply chan string) {
	m := &S_uplinkmsg{}
	err := json.Unmarshal([]byte(req), m)
	if err != nil {
		log.Println("uplinkmsg Unmarshal err", err)
		return
	}
	meta := py.NewDict()
	defer meta.Decref()

	msgfrom := py.NewString(m.Msgfrom)
	defer msgfrom.Decref()
	meta.SetItemString("msgfrom", msgfrom.Obj())

	action := py.NewString(m.Action)
	defer action.Decref()
	meta.SetItemString("action", action.Obj())

	appid := py.NewInt64(m.Appid)
	defer appid.Decref()
	meta.SetItemString("appid", appid.Obj())

	suid := py.NewInt64(m.Suid)
	defer suid.Decref()
	meta.SetItemString("suid", suid.Obj())

	uid := py.NewInt64(m.Uid)
	defer uid.Decref()
	meta.SetItemString("uid", uid.Obj())

	topsid := py.NewInt64(m.Topsid)
	defer topsid.Decref()
	meta.SetItemString("topsid", topsid.Obj())

	subsid := py.NewInt64(m.Subsid)
	defer subsid.Decref()
	meta.SetItemString("subsid", subsid.Obj())

	yyfrom := py.NewString(m.Yyfrom)
	defer yyfrom.Decref()
	meta.SetItemString("yyfrom", yyfrom.Obj())

	terminaltype := py.NewInt64(m.Terminaltype)
	defer terminaltype.Decref()
	meta.SetItemString("terminaltype", terminaltype.Obj())

	userip := py.NewString(m.Userip)
	defer userip.Decref()
	meta.SetItemString("userip", userip.Obj())

	data := py.NewString(m.Data)
	defer data.Decref()

	if _, err := this.caller.callPyFunc("OnUplinkmsg", meta.Obj(), data.Obj()); err != nil {
		log.Println("OnUplinkmsg err:", err)
	}
	reply <- ""
}

type S_leaveplatform struct {
	Msgfrom string `json:"data"`
	Action  string `json:"data"`
	Suid    int64  `json:"data"`
	Uid     int64  `json:"data"`
}

func (this *ServiceModule) leaveplatform(req string, reply chan string) {
	m := &S_leaveplatform{}
	err := json.Unmarshal([]byte(req), m)
	if err != nil {
		log.Println("leaveplatform Unmarshal err", err)
		return
	}
	meta := py.NewDict()
	defer meta.Decref()

	msgfrom := py.NewString(m.Msgfrom)
	defer msgfrom.Decref()
	meta.SetItemString("msgfrom", msgfrom.Obj())

	action := py.NewString(m.Action)
	defer action.Decref()
	meta.SetItemString("action", action.Obj())

	suid := py.NewInt64(m.Suid)
	defer suid.Decref()
	meta.SetItemString("suid", suid.Obj())

	uid := py.NewInt64(m.Uid)
	defer uid.Decref()

	if _, err := this.caller.callPyFunc("OnLeaveplatform", meta.Obj(), uid.Obj()); err != nil {
		log.Println("OnUplinkmsg err:", err)
	}
	reply <- ""
}
