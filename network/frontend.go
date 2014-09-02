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
	cc       *ClientConnection
	fid      uint32
	buffChan chan *ConnBuff

	GateInChan  chan *proto.GateInPack
	GateOutChan chan *proto.GateOutPack
}

func NewFrontGate(entry chan *proto.GateInPack, exit chan *proto.GateOutPack) *FrontGate {
	conn, err := net.Dial("tcp", conf.CF.DIAL_MASTER_ADDR)
	if err != nil {
		log.Fatalln("dial to master err", err)
	}
	fe := &FrontGate{fid: uint32(conf.CF.FID), cc: NewClientConnection(conn),
		buffChan:    make(chan *ConnBuff, conf.CF.BUF_QUEUE),
		GateInChan:  entry,
		GateOutChan: exit}
	return fe
}

func (this *FrontGate) Start() {
	log.Println("FrontGate", this.fid, "running, dial to", conf.CF.DIAL_MASTER_ADDR)
	this.register()
	go this.comein()
	go this.comeout()
}

func (this *FrontGate) register() {
	pack := &proto.FrontendRegister{Fid: pb.Uint32(this.fid)}
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

// socket buff -> FrontendPack
func (this *FrontGate) comein() {
	for conn_buff := range this.buffChan {
		msg := conn_buff.buff
		if msg == nil {
			log.Println("disconnect")
			return
		}
		fp := &proto.FrontendPack{}
		if err := pb.Unmarshal(msg[LEN_URI:], fp); err == nil {
			this.GateInChan <- &proto.GateInPack{Tsid: fp.Tsid,
				Ssid: fp.Ssid, Uri: fp.Uri, Bin: fp.Bin, Uid: fp.Uid}
		} else {
			log.Println("Unmarshal FrontGatePack err")
		}
	}
}

// GateOutPack -> FrontendPack
// unicast only
func (this *FrontGate) comeout() {
	for pack := range this.GateOutChan {
		fp := &proto.FrontendPack{Uri: pack.Uri, Tsid: pack.Tsid, Ssid: pack.Ssid,
			Bin: pack.Bin, Fid: pb.Uint32(this.fid), Uid: pack.Uid}
		if data, err := pb.Marshal(fp); err == nil {
			uri_field := make([]byte, LEN_URI)
			binary.LittleEndian.PutUint32(uri_field, uint32(URI_TRANSPORT))
			bin := append(uri_field, data...)
			this.cc.Send(bin)
		} else {
			log.Println("[Error]pack FrontendPack", err)
		}
	}
}
