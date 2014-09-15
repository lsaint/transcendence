package network

import (
	"encoding/binary"
	"log"
	"net"
	"transcendence/conf"
	"transcendence/proto"

	pb "code.google.com/p/goprotobuf/proto"
)

const (
	NODE_MASTER = 1
	NODE_DRONE  = 2
)

// for connect
type FrontGate struct {
	cc       *IConnection
	fid      uint32
	buffChan chan *ConnBuff

	GateInChan  chan *proto.Passpack
	GateOutChan chan *proto.Passpack
}

func NewFrontGate(entry chan *proto.Passpack, exit chan *proto.Passpack) *FrontGate {
	conn, err := net.Dial("tcp", conf.CF.DIAL_HIVE_ADDR)
	if err != nil {
		log.Fatalln("dial to master err", err)
	}
	fe := &FrontGate{fid: uint32(conf.CF.FID), cc: NewIConnection(conn),
		buffChan:    make(chan *ConnBuff, conf.CF.BUF_QUEUE),
		GateInChan:  entry,
		GateOutChan: exit}
	return fe
}

func (this *FrontGate) Start() {
	log.Println("FrontGate", this.fid, "running, dial to", conf.CF.DIAL_HIVE_ADDR)
	this.register()
	go this.comein()
	go this.comeout()
}

func (this *FrontGate) register() {
	pack := &proto.Passpack{Tsid: pb.Uint32(0), Ssid: pb.Uint32(0), Uri: pb.Uint32(0),
		Bin: nil, Fid: pb.Uint32(this.fid), Action: proto.Action_D2H_Register.Enum()}
	if data, err := pb.Marshal(pack); err == nil {
		uri_field := make([]byte, LEN_URI)
		binary.LittleEndian.PutUint32(uri_field, uint32(URI_REGISTER))
		bin := append(uri_field, data...)
		this.cc.Send(bin)
	} else {
		log.Println("[Error]marshal register pack err")
	}
}

func (this *FrontGate) recvFromSock() {
	for {
		if buff_body, ok := this.cc.ReadBody(); ok {
			this.buffChan <- &ConnBuff{nil, buff_body}
			continue
		}
		this.buffChan <- &ConnBuff{nil, nil}
		break
	}

}

// socket buff -> Passpack
func (this *FrontGate) comein() {
	for conn_buff := range this.buffChan {
		msg := conn_buff.buff
		if msg == nil {
			log.Println("disconnect")
			return
		}
		p := &proto.Passpack{}
		if err := pb.Unmarshal(msg[LEN_URI:], p); err == nil {
			this.GateInChan <- p
		} else {
			log.Println("Unmarshal Passpack err")
		}
	}
}

// unicast only
func (this *FrontGate) comeout() {
	for pack := range this.GateOutChan {
		pack.Fid = pb.Uint32(this.fid)
		if data, err := pb.Marshal(pack); err == nil {
			uri_field := make([]byte, LEN_URI)
			binary.LittleEndian.PutUint32(uri_field, pack.GetUri())
			bin := append(uri_field, data...)
			this.cc.Send(bin)
		} else {
			log.Println("[Error]pack FrontendPack", err)
		}
	}
}
