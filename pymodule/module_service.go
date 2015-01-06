package pymodule

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
	"transcendence/network"

	"github.com/lsaint/py"

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
	*ClientMsgBroker
}

func NewServiceModule(caller PyFuncCaller) *ServiceModule {
	httpChan := make(chan *network.HttpReq, I("BUF_QUEUE"))
	service := &ServiceModule{
		pm:       network.NewPostman(),
		httpChan: httpChan,
		caller:   caller,
		httpSrv: network.NewHttpServer(httpChan, fmt.Sprintf(":%v", I("SERVICE_LISTEN_PORT")),
			[]string{"/uplinkmsg", "/leaveplatform", "/checkalive"}),
		ClientMsgBroker: NewClientMsgBroker(),
	}
	service.init()
	return service
}

func (this *ServiceModule) init() {
	go this.httpSrv.Start()
	this.register()
	go this.processServiceMsg()
	go this.proto2py()
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
	m, err := url.ParseQuery(req)
	if err != nil {
		log.Println("uplinkmsg ParseQuery err", err)
		return
	}

	this.OnUplinkmsg(m)
	reply <- ""
}

func (this *ServiceModule) proto2py() {
	for p := range this.ClientProtoChan {
		meta := py.NewDict()
		defer meta.Decref()

		msgfrom := py.NewString(p.Msgfrom)
		defer msgfrom.Decref()
		meta.SetItemString("msgfrom", msgfrom.Obj())

		action := py.NewString(p.Action)
		defer action.Decref()
		meta.SetItemString("action", action.Obj())

		appid := py.NewInt64(p.Appid)
		defer appid.Decref()
		meta.SetItemString("appid", appid.Obj())

		suid := py.NewInt64(p.Suid)
		defer suid.Decref()
		meta.SetItemString("suid", suid.Obj())

		uid := py.NewInt64(p.Uid)
		defer uid.Decref()
		meta.SetItemString("uid", uid.Obj())

		topsid := py.NewInt64(p.Topsid)
		defer topsid.Decref()
		meta.SetItemString("topsid", topsid.Obj())

		subsid := py.NewInt64(p.Subsid)
		defer subsid.Decref()
		meta.SetItemString("subsid", subsid.Obj())

		yyfrom := py.NewString(p.Yyfrom)
		defer yyfrom.Decref()
		meta.SetItemString("yyfrom", yyfrom.Obj())

		terminaltype := py.NewInt64(p.Terminaltype)
		defer terminaltype.Decref()
		meta.SetItemString("terminaltype", terminaltype.Obj())

		userip := py.NewString(p.Userip)
		defer userip.Decref()
		meta.SetItemString("userip", userip.Obj())

		data := py.NewStringWithSize(p.Data, len(p.Data))
		defer data.Decref()

		uri := py.NewInt64(p.Uri)
		defer uri.Decref()

		if _, err := this.caller.callPyFunc("OnUplinkmsg",
			uri.Obj(), meta.Obj(), data.Obj()); err != nil {
			log.Println("OnUplinkmsg err:", err)
		}
	}
}

type S_leaveplatform struct {
	Msgfrom string `json:"msgfrom"`
	Action  string `json:"action"`
	Suid    int64  `json:"suid"`
	Uid     int64  `json:"uid"`
}

func (this *ServiceModule) leaveplatform(req string, reply chan string) {
	m, err := url.ParseQuery(req)
	if err != nil {
		log.Println("leaveplatform ParseQuery err", err)
		return
	}
	meta := py.NewDict()
	defer meta.Decref()

	msgfrom := py.NewString(m["msgfrom"][0])
	defer msgfrom.Decref()
	meta.SetItemString("msgfrom", msgfrom.Obj())

	action := py.NewString(m["action"][0])
	defer action.Decref()
	meta.SetItemString("action", action.Obj())

	_suid, err := strconv.Atoi(m["suid"][0])
	suid := py.NewInt64(int64(_suid))
	defer suid.Decref()
	meta.SetItemString("suid", suid.Obj())

	_uid, err := strconv.Atoi(m["uid"][0])
	uid := py.NewInt64(int64(_uid))
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

//

type ClientMsgBroker struct {
	uid2clientbuff  map[uint32]*ClientBuff
	tmpbuf          []byte
	ClientProtoChan chan *ClientProto
}

func NewClientMsgBroker() *ClientMsgBroker {
	return &ClientMsgBroker{uid2clientbuff: make(map[uint32]*ClientBuff),
		tmpbuf:          make([]byte, 2),
		ClientProtoChan: make(chan *ClientProto, I("BUF_QUEUE"))}
}

func (this *ClientMsgBroker) OnUplinkmsg(m url.Values) {
	uid, _ := strconv.Atoi(m["uid"][0])
	subsid, _ := strconv.Atoi(m["subsid"][0])
	cbuff, exist := this.uid2clientbuff[uint32(uid)]
	if !exist {
		cbuff := NewClientBuff(uint32(uid), uint32(subsid), m)
		go this.acceptConn(cbuff)
	} else if cbuff.Subsid != uint32(subsid) {
		cbuff.Reset()
	}
	cbuff.WriteString(m["data"][0])
	m.Del("data")
}

func (this *ClientMsgBroker) acceptConn(cbuff *ClientBuff) {
	c := network.NewIConnection(cbuff)
	for {
		if buff_body, ok := c.ReadBody(); ok {
			uri := binary.LittleEndian.Uint32(buff_body[:network.LEN_URI])
			cbuff.Read(this.tmpbuf)
			m := cbuff.meta

			appid, _ := strconv.Atoi(m["appid"][0])
			suid, _ := strconv.Atoi(m["suid"][0])
			uid, _ := strconv.Atoi(m["uid"][0])
			subsid, _ := strconv.Atoi(m["subsid"][0])
			topsid, _ := strconv.Atoi(m["topsid"][0])
			terminaltype, _ := strconv.Atoi(m["terminaltype"][0])

			this.ClientProtoChan <- &ClientProto{
				Msgfrom:      m["msgfrom"][0],
				Action:       m["action"][0],
				Appid:        int64(appid),
				Suid:         int64(suid),
				Uid:          int64(uid),
				Topsid:       int64(topsid),
				Subsid:       int64(subsid),
				Yyfrom:       m["yyfrom"][0],
				Terminaltype: int64(terminaltype),
				Userip:       m["userip"][0],
				Data:         string(buff_body[10:]), // len + magic + uri = 10
				Uri:          int64(uri)}
		}
	}
	c.Close()
}

type ClientProto struct {
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

	Uri int64 `json:"uri"`
}

type ClientBuff struct {
	Subsid uint32
	Uid    uint32
	meta   url.Values
	bytes.Buffer
}

func NewClientBuff(uid, ssid uint32, meta url.Values) *ClientBuff {
	var b bytes.Buffer
	return &ClientBuff{Uid: uid, Subsid: ssid, meta: meta, Buffer: b}
}

func (this *ClientBuff) Close() error {
	this.Reset()
	return nil
}
