package pymodule

import (
	"log"
	"thrift/salService"
	"transcendence/conf"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/qiniu/py"
)

type SalModule struct {
	glue *py.Module

	protocolFactory  thrift.TProtocolFactory
	transportFactory thrift.TTransportFactory

	SalClient *salService.SalServiceClient
	SalServer *thrift.TSimpleServer
	SvcType   int32
}

func NewSalModule(svctype int, glue *py.Module) *SalModule {
	mod := &SalModule{}
	mod.SvcType = int32(svctype)
	mod.glue = glue
	mod.init()
	mod.runSalClient()
	go mod.runSalServer()
	return mod
}

func (this *SalModule) init() {
	this.protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	this.transportFactory = thrift.NewTBufferedTransportFactory(conf.CF.BUF_QUEUE)
	this.transportFactory = thrift.NewTFramedTransportFactory(this.transportFactory)
}

func (this *SalModule) runSalClient() {
	var transport thrift.TTransport
	transport, err := thrift.NewTSocket(conf.CF.SAL_REMOTE_ADDR)
	if err != nil {
		log.Fatalln("connect to sal err", err)
	}
	transport = this.transportFactory.GetTransport(transport)

	if err = transport.Open(); err != nil {
		log.Fatalln("sal transport Open err:", err)
	}

	this.SalClient = salService.NewSalServiceClientFactory(transport, this.protocolFactory)
	r, err := this.SalClient.SALLogin(this.SvcType)
	if err != nil {
		log.Fatalln("SALLogin FAIL", err)
	}
	log.Println("SALLogin sucess", r)
}

func (this *SalModule) runSalServer() {
	transport, err := thrift.NewTServerSocket(conf.CF.SAL_LOCAL_ADDR)
	if err != nil {
		log.Fatalln("NewTServerSocket err", err)
	}

	handler := NewSalLocalServerHandler(this.glue)
	processor := salService.NewSalServiceProcessor(handler)

	this.SalServer = thrift.NewTSimpleServer4(processor, transport,
		this.transportFactory, this.protocolFactory)
	log.Println("local server running")
	this.SalServer.Serve()
}

//
func (this *SalModule) Py_SALPing(args *py.Tuple) (ret *py.Base, err error) {
	if _, err := this.SalClient.SALPing(this.SvcType); err == nil {
		return py.NewInt(1).Obj(), nil
	} else {
		return py.IncNone(), nil
	}
}

func (this *SalModule) Py_SALMsgToClient(args *py.Tuple) (ret *py.Base, err error) {
	var data string
	var topSid, uid int
	err = py.Parse(args, &topSid, &uid, &data)
	if err != nil {
		return py.IncNone(), err
	}
	r, err := this.SalClient.SALMsgToClient(this.SvcType, int64(topSid), int64(uid), data)
	if err != nil {
		return py.IncNone(), err
	}
	return py.NewInt(int(r)).Obj(), nil
}

func (this *SalModule) Py_SALTopSidBroadcast(args *py.Tuple) (ret *py.Base, err error) {
	var data string
	var topSid, exceptUid int
	err = py.Parse(args, &topSid, &exceptUid, &data)
	if err != nil {
		return py.IncNone(), err
	}
	r, err := this.SalClient.SALTopSidBroadcast(this.SvcType, int64(topSid), int64(exceptUid), data)
	if err != nil {
		return py.IncNone(), err
	}
	return py.NewInt(int(r)).Obj(), nil
}

func (this *SalModule) Py_SALSubSidBroadcast(args *py.Tuple) (ret *py.Base, err error) {
	var data string
	var topSid, subSid, exceptUid int
	err = py.Parse(args, &topSid, &subSid, &exceptUid, &data)
	if err != nil {
		return py.IncNone(), err
	}
	r, err := this.SalClient.SALSubSidBroadcast(this.SvcType, int64(topSid),
		int64(subSid), int64(exceptUid), data)
	if err != nil {
		return py.IncNone(), err
	}
	return py.NewInt(int(r)).Obj(), nil
}

func (this *SalModule) Py_SALMulticast(args *py.Tuple) (ret *py.Base, err error) {
	var data string
	var topSid, operatorUid, groupId int
	var uids []int
	err = py.ParseV(args, &topSid, &operatorUid, &groupId, &data, &uids)
	if err != nil {
		return py.IncNone(), err
	}

	uidSet := make(map[int32]bool)
	for uid := range uids {
		uidSet[int32(uid)] = true
	}
	r, err := this.SalClient.SALMulticast(this.SvcType, int64(topSid),
		int32(operatorUid), int64(groupId), data, uidSet)
	if err != nil {
		return py.IncNone(), err
	}
	return py.NewInt(int(r)).Obj(), nil
}

func (this *SalModule) Py_SALMulticast2(args *py.Tuple) (ret *py.Base, err error) {
	var data string
	var topSid, multicastGroupId int
	var uids []int
	err = py.ParseV(args, &topSid, &multicastGroupId, &data, &uids)
	if err != nil {
		return py.IncNone(), err
	}

	uidSet := make(map[int32]bool)
	for uid := range uids {
		uidSet[int32(uid)] = true
	}
	r, err := this.SalClient.SALMulticast2(this.SvcType, int64(topSid),
		int64(multicastGroupId), data, uidSet)
	if err != nil {
		return py.IncNone(), err
	}
	return py.NewInt(int(r)).Obj(), nil
}

func (this *SalModule) Py_SALSubscribeHashChannelRange(args *py.Tuple) (ret *py.Base, err error) {
	var topSid, subStart, subEnd int
	err = py.Parse(args, &topSid, &subStart, &subEnd)
	if err != nil {
		return py.IncNone(), err
	}
	r, err := this.SalClient.SALSubscribeHashChannelRange(this.SvcType, int64(topSid),
		int32(subStart), int32(subEnd))
	if err != nil {
		return py.IncNone(), err
	}
	return py.NewInt(int(r)).Obj(), nil
}

func (this *SalModule) Py_SALUnicast(args *py.Tuple) (ret *py.Base, err error) {
	var data string
	var topSid, uid int
	err = py.Parse(args, &topSid, &uid, &data)
	if err != nil {
		return py.IncNone(), err
	}
	r, err := this.SalClient.SALUnicast(this.SvcType, int64(topSid), int64(uid), data)
	if err != nil {
		return py.IncNone(), err
	}
	return py.NewInt(int(r)).Obj(), nil
}

func (this *SalModule) Py_SALSSubscribeUserInOutMove(args *py.Tuple) (ret *py.Base, err error) {
	var tids []int
	err = py.ParseV(args, &tids)
	if err != nil {
		return py.IncNone(), err
	}

	tidSet := make(map[int32]bool)
	for tid := range tids {
		tidSet[int32(tid)] = true
	}
	log.Println("inout tids", tids)
	r, err := this.SalClient.SALSSubscribeUserInOutMove(this.SvcType, tidSet)
	if err != nil {
		return py.IncNone(), err
	}
	return py.NewInt(int(len(r))).Obj(), nil
}

func (this *SalModule) Py_SALSSubscribeMaixuQueueChange(args *py.Tuple) (ret *py.Base, err error) {
	var tids []int
	err = py.ParseV(args, &tids)
	if err != nil {
		return py.IncNone(), err
	}

	tidSet := make(map[int32]bool)
	for tid := range tids {
		tidSet[int32(tid)] = true
	}
	r, err := this.SalClient.SALSSubscribeMaixuQueueChange(this.SvcType, tidSet)
	if err != nil {
		return py.IncNone(), err
	}
	return py.NewInt(int(len(r))).Obj(), nil
}

func (this *SalModule) Py_SALSQueryMaixuQueue(args *py.Tuple) (ret *py.Base, err error) {
	var topSid, subSid int
	err = py.Parse(args, &topSid, &subSid)
	if err != nil {
		return py.IncNone(), err
	}
	r, err := this.SalClient.SALSQueryMaixuQueue(this.SvcType, int64(topSid), int64(subSid))
	if err != nil {
		return py.IncNone(), err
	}
	return py.NewInt(int(r)).Obj(), nil
}

func (this *SalModule) Py_SALSQueryUserRole(args *py.Tuple) (ret *py.Base, err error) {
	var topSid, uid int
	err = py.Parse(args, &topSid, &uid)
	if err != nil {
		return py.IncNone(), err
	}
	r, err := this.SalClient.SALSQueryUserRole(this.SvcType, int64(topSid), int64(uid))
	if err != nil {
		return py.IncNone(), err
	}
	return py.NewInt(int(r)).Obj(), nil
}

func (this *SalModule) Py_SALSocketInfo(args *py.Tuple) (ret *py.Base, err error) {
	var ip string
	var port int
	err = py.Parse(args, &ip, &port)
	if err != nil {
		return py.IncNone(), err
	}
	r, err := this.SalClient.SALSocketInfo(this.SvcType, ip, int32(port))
	if err != nil {
		return py.IncNone(), err
	}
	return py.NewInt(int(r)).Obj(), nil
}

// sal local server handler
type SalLocalServerHandler struct {
	glue *py.Module
}

func NewSalLocalServerHandler(glue *py.Module) *SalLocalServerHandler {
	return &SalLocalServerHandler{glue}
}

func (this *SalLocalServerHandler) SALPing(SvcType int32) (r bool, err error) {
	return
}
func (this *SalLocalServerHandler) SALLogin(SvcType int32) (r string, err error) {
	return
}
func (this *SalLocalServerHandler) SALMsgToClient(SvcType int32, topSid int64, uid int64, data string) (r int32, err error) {
	return
}
func (this *SalLocalServerHandler) SALTopSidBroadcast(SvcType int32, topSid int64, exceptUid int64, data string) (r int32, err error) {
	return
}
func (this *SalLocalServerHandler) SALSubSidBroadcast(SvcType int32, topSid int64, subSid int64, exceptUid int64, data string) (r int32, err error) {
	return
}
func (this *SalLocalServerHandler) SALMulticast(SvcType int32, topSid int64, operatorUid int32, groupId int64, data string, uidSet map[int32]bool) (r int32, err error) {
	return
}
func (this *SalLocalServerHandler) SALMulticast2(SvcType int32, topSid int64, multicastGroupId int64, data string, uidSet map[int32]bool) (r int32, err error) {
	return
}
func (this *SalLocalServerHandler) SALSubscribeHashChannelRange(SvcType int32, topSid int64, subStart int32, subEnd int32) (r int32, err error) {
	return
}
func (this *SalLocalServerHandler) SALUnicast(SvcType int32, topSid int64, uid int64, data string) (r int32, err error) {
	return
}
func (this *SalLocalServerHandler) SALSSubscribeUserInOutMove(SvcType int32, tidSet map[int32]bool) (r []int32, err error) {
	return
}
func (this *SalLocalServerHandler) SALSSubscribeMaixuQueueChange(SvcType int32, tidSet map[int32]bool) (r []int32, err error) {
	return
}
func (this *SalLocalServerHandler) SALSQueryMaixuQueue(SvcType int32, topSid int64, subSid int64) (r int32, err error) {
	return
}
func (this *SalLocalServerHandler) SALSQueryUserRole(SvcType int32, topSid int64, uid int64) (r int32, err error) {
	return
}
func (this *SalLocalServerHandler) SALSocketInfo(SvcType int32, ip string, port int32) (r int32, err error) {
	return
}

//

func (this *SalLocalServerHandler) SALCPing(SvcType int32) (r bool, err error) {
	log.Println("->SALCPing")
	return true, nil
}
func (this *SalLocalServerHandler) SALSubscribeUserInOutMove(SvcType int32, return_ []*salService.UserInOut) (r int32, err error) {
	log.Println("->SALSubscribeUserInOutMove", return_)
	tp := py.NewTuple(len(return_))
	for i, item := range return_ {
		m := py.NewDict()
		flag := py.NewInt64(int64(item.Flag))
		defer flag.Decref()
		m.SetItemString("Flag", flag.Obj())
		name := py.NewString(item.Name)
		defer name.Decref()
		m.SetItemString("Name", name.Obj())
		uid := py.NewInt64(int64(item.Uid))
		defer uid.Decref()
		m.SetItemString("Uid", uid.Obj())
		sid := py.NewInt64(int64(item.Sid))
		defer sid.Decref()
		m.SetItemString("Sid", sid.Obj())
		dsid := py.NewInt64(int64(item.Dsid))
		defer dsid.Decref()
		m.SetItemString("Dsid", dsid.Obj())
		tp.SetItem(i, m.Obj())
	}
	_, err = this.glue.CallMethodObjArgs("SALSubscribeUserInOutMove", tp.Obj())
	if err != nil {
		log.Println("py call SALSubscribeUserInOutMove err:", err)
	}
	return
}
func (this *SalLocalServerHandler) SALSubscribeMaixuQueueChange(SvcType int32, return_ *salService.SubscribeMX) (r int32, err error) {
	log.Println("->SALSubscribeMaixuQueueChange", return_)
	return
}
func (this *SalLocalServerHandler) SALQueryMaixuQueue(SvcType int32, return_ map[int32][]int32) (r int32, err error) {
	log.Println("->SALQueryMaixuQueue", return_)
	return
}
func (this *SalLocalServerHandler) SALQueryUserRole(SvcType int32, return_ []*salService.QueueUserRole) (r int32, err error) {
	log.Println("->SALQueryUserRole", return_)
	return
}
func (this *SalLocalServerHandler) SALMsgFromClient(SvcType int32, return_ *salService.MsgFromClient) (r int32, err error) {
	log.Println("->SALMsgFromClient", return_)
	m := py.NewDict()
	flag := py.NewInt64(int64(return_.Flag))
	defer flag.Decref()
	m.SetItemString("Flag", flag.Obj())
	topsid := py.NewInt64(int64(return_.TopSid))
	defer topsid.Decref()
	m.SetItemString("TopSid", topsid.Obj())
	uid := py.NewInt64(int64(return_.Uid))
	defer uid.Decref()
	m.SetItemString("Uid", uid.Obj())
	msg := py.NewString(return_.Msg)
	defer msg.Decref()
	m.SetItemString("Msg", msg.Obj())
	_, err = this.glue.CallMethodObjArgs("SALMsgFromClient", m.Obj())
	if err != nil {
		log.Println("py call SALMsgFromClient err:", err)
	}
	return
}
