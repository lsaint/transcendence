package pymodule

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"
	"transcendence/network"

	"github.com/qiniu/py"

	. "transcendence/conf"
)

type PyFuncCaller interface {
	callPyFunc(string, ...*py.Base) (*py.Base, error)
}

type ServiceModule struct {
	pm       *network.Postman
	httpChan chan *network.HttpReq
	httpSrv  *network.HttpServer
	caller   PyFuncCaller
}

func NewServiceModule(caller PyFuncCaller) *ServiceModule {
	httpChan := make(chan *network.HttpReq, I("BUF_QUEUE"))
	service := &ServiceModule{
		pm:       network.NewPostman(),
		httpChan: httpChan,
		caller:   caller,
		httpSrv: network.NewHttpServer(httpChan, fmt.Sprintf(":%v", I("SERVICE_LISTEN_PORT")),
			[]string{"/uplinkmsg", "/leaveplatform", "/checkalive"}),
	}
	service.init()
	return service
}

func (this *ServiceModule) init() {
	go this.httpSrv.Start()
	this.register()
	go this.processServiceMsg()
}

/*
CTX_REG = {"appid":"120","appname":"testreghttp","port":"20000","groupid":"4500","isp":"3",
	"isp2ip":[{"ip":"121.14.170.11","isp":"2"}], "regkey":"1c4814e155bd8e0acdcc3d5748e952de"}
*/
func (this *ServiceModule) register() {
	ret := this.pm.Post(S("URL_SERVICE_REG"), S("CTX_REG"))
	if ret == "" {
		log.Fatalln("register to service timeout")
	}
	log.Println("[SERVICE]register sucess", ret)
}

func (this *ServiceModule) processServiceMsg() {
	for req := range this.httpChan {
		switch req.Url {
		case "/uplinkmsg":
			this.uplinkmsg(req.Req, req.Ret)
		case "/leaveplatform":
			this.leaveplatform(req.Req, req.Ret)
		case "/checkalive":
			this.checkalive(req)
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
	Msgfrom string `json:"msgfrom"`
	Action  string `json:"action"`
	Suid    int64  `json:"suid"`
	Uid     int64  `json:"uid"`
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

func (this *ServiceModule) checkalive(r *network.HttpReq) {
	log.Println("http-service checkalive")
	r.Ret <- fmt.Sprintf("response=checkalive&ts=%v", time.Now().Unix())
}

type castResp struct {
	Response string `json:"response"`
	Message  string `json:"message"`
}

func (this *ServiceModule) doCast(url, body string) {
	resp := &castResp{}
	err := json.Unmarshal([]byte(this.pm.Post(url, body)), resp)
	if err != nil || resp.Message != "OK" {
		log.Println("cast err", err, resp)
	}
}

/*
appid	必需，应用申请的appid(service type)
regkey	必需，应用申请appid时对应的regkey
uid		必需，单播的目标uid
topsid	可选，不为0时，仅当目标uid在对应的频道时才将消息投递给用户
Request body

URL: http://abc/unicast?appid=120&regkey=1c4814e155bd8e0a26d69c45bd024531&uid=888999&topsid=400000
POST内容：hello world
*/
func (this *ServiceModule) Py_Unicast(args *py.Tuple) (ret *py.Base, err error) {
	var uid, topsid int64
	var body string
	err = py.Parse(args, &body, &topsid, &uid)
	if err != nil {
		log.Println("Parse unicast err:", err)
		return
	}

	subfix := url.Values{}
	subfix.Set("appid", fmt.Sprintf("%v", I("SERVICE_APPID")))
	subfix.Add("reqkey", S("SERVICE_REGKEY"))
	subfix.Add("uid", fmt.Sprintf("%v", uid))
	subfix.Add("topsid", fmt.Sprintf("%v", topsid))

	u := fmt.Sprintf("%v/%v", S("URL_SERVICE_UNICAST"), subfix.Encode())
	go this.doCast(u, body)

	return py.IncNone(), nil
}

/*
appid	必需，应用申请的appid(service type)
regkey	必需，应用申请appid时对应的regkey
uids	多播的目标uid列表，以逗号分隔，如”80001,80002,80003”等等
topsid	可选，不为0时，仅当目标uid在对应的频道时才将消息投递给用户
Request body	POST方法时提交的消息内容，亦即客户端将收到的内容

URL: http://abc/multicast
		?appid=120&regkey=1c4814e155bd8e0a26d69c45bd024531&uids=80001,80002,80003&topsid=400000
POST内容：hello world
*/
func (this *ServiceModule) Py_Multicast(args *py.Tuple) (ret *py.Base, err error) {
	var uids []int
	var topsid int
	var body string
	err = py.ParseV(args, &body, &topsid, &uids)
	if err != nil {
		log.Println("Parse unicast err:", err)
		return
	}

	s_uids := make([]string, len(uids))
	for i, uid := range uids {
		s_uids[i] = fmt.Sprintf("%v", uid)
	}

	subfix := url.Values{}
	subfix.Set("appid", fmt.Sprintf("%v", I("SERVICE_APPID")))
	subfix.Add("reqkey", S("SERVICE_REGKEY"))
	subfix.Add("topsid", fmt.Sprintf("%v", topsid))
	subfix.Add("uids", strings.Join(s_uids, ","))

	u := fmt.Sprintf("%v/%v", S("URL_SERVICE_MULTICAST"), subfix.Encode())
	go this.doCast(u, body)

	return py.IncNone(), nil
}

/*
appid	必需，应用申请的appid(service type)
regkey	必需，应用申请appid时对应的regkey
topsid	可选，当topsid不为0而subsid为0时，topsid指定的频道内的所有人将收到消息（包括子频道）
subsid	.......
group_type	可选，与group_id共同组成应用自定义的广播类型（参考服务端SDK的说明）
group_id	可选，与goup_type共同组成应用自定义的广播类型（参考服务端SDK的说明）
op_uid	可选，当使用group_type和group_id时，op_uid指求发起广播请求的用户
Request body	POST方法时提交的消息内容，亦即客户端将收到的内容

URL: http://abc/broadcast?appid=120&regkey=1c4814e155bd8e0a26d69c45bd024531&topsid=400000
POST内容：hello world
*/
func (this *ServiceModule) Py_Broadcast(args *py.Tuple) (ret *py.Base, err error) {
	var topsid, subsid int
	var body string

	err = py.Parse(args, &body, &topsid, &subsid)
	if err != nil {
		log.Println("Parse unicast err:", err)
		return
	}

	subfix := url.Values{}
	subfix.Set("appid", fmt.Sprintf("%v", I("SERVICE_APPID")))
	subfix.Add("reqkey", S("SERVICE_REGKEY"))
	subfix.Add("topsid", fmt.Sprintf("%v", topsid))
	subfix.Add("subsid", fmt.Sprintf("%v", subsid))

	u := fmt.Sprintf("%v/%v", S("URL_SERVICE_BROADCAST"), subfix.Encode())
	go this.doCast(u, body)

	return py.IncNone(), nil
}
