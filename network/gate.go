package network

import (
	"encoding/binary"
	"log"
	"math/rand"
	"net"

	pb "code.google.com/p/goprotobuf/proto"

	"transcendence/conf"
	"transcendence/proto"
)

const (
	URI_REGISTER   = 1
	URI_TRANSPORT  = 2
	URI_UNREGISTER = 3
	URI_PING       = 4
)

// for accept
type BackGate struct {
	buffChan     chan *ConnBuff
	fid2frontend map[uint32]*ClientConnection
	uid2fid      map[uint32]uint32
	fids         []uint32
	GateInChan   chan *proto.GateInPack
	GateOutChan  chan *proto.GateOutPack
}

func NewBackGate(entry chan *proto.GateInPack, exit chan *proto.GateOutPack) *BackGate {
	gs := &BackGate{buffChan: make(chan *ConnBuff, conf.CF.BUF_QUEUE),
		fid2frontend: make(map[uint32]*ClientConnection),
		uid2fid:      make(map[uint32]uint32),
		fids:         make([]uint32, 0),
		GateInChan:   entry,
		GateOutChan:  exit}
	go gs.parse()
	return gs
}

func (this *BackGate) Start() {
	ln, err := net.Listen("tcp", conf.CF.SRV_ADDR)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("[Info]BackGate running", conf.CF.SRV_ADDR)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("[Error]Accept", err)
			continue
		}
		log.Println("[Info]frontend connected")
		go this.acceptConn(conn)
	}
}

func (this *BackGate) acceptConn(conn net.Conn) {
	cliConn := NewClientConnection(conn)
	for {
		if buff_body, ok := cliConn.ReadBody(); ok {
			this.buffChan <- &ConnBuff{cliConn, buff_body}
			continue
		}
		this.buffChan <- &ConnBuff{cliConn, nil}
		break
	}
	log.Println("[Info]frontend disconnect")
	cliConn.Close()
}

func (this *BackGate) parse() {
	go func() {
		for pack := range this.GateOutChan {
			this.comeout(pack)
		}
	}()

	for conn_buff := range this.buffChan {
		msg := conn_buff.buff
		conn := conn_buff.conn
		if msg == nil {
			this.unregister(conn)
			continue
		}
		len_msg := len(msg)
		if len_msg < LEN_URI {
			continue
		}

		f_uri := binary.LittleEndian.Uint32(msg[:LEN_URI])
		switch f_uri {
		case URI_REGISTER:
			this.register(msg[LEN_URI:], conn)
		case URI_TRANSPORT:
			this.comein(msg[LEN_URI:])
		case URI_UNREGISTER:
			this.unregister(conn)
		case URI_PING:

		default:
			log.Println("[Error]invalid f_uri:", f_uri)
		}
	}
}

func (this *BackGate) unpack(b []byte) (msg *proto.GateInPack, err error) {
	fp := &proto.FrontendPack{}
	if err = pb.Unmarshal(b, fp); err == nil {
		// register uid2fid
		if fp.GetUid() != 0 {
			this.uid2fid[fp.GetUid()] = fp.GetFid()
		}
		msg = &proto.GateInPack{Tsid: fp.Tsid, Ssid: fp.Ssid, Uri: fp.Uri, Bin: fp.Bin, Uid: fp.Uid}
	} else {
		log.Println("[Error]pb Unmarshal FrontendPack", err)
	}
	return
}

func (this *BackGate) comein(b []byte) {
	if msg, err := this.unpack(b); err == nil {
		this.GateInChan <- msg
	}
}

func (this *BackGate) randomFid() uint32 {
	return this.fids[rand.Intn(len(this.fids))]
}

func (this *BackGate) broadcastFid() uint32 {
	return this.fids[0]
}

func (this *BackGate) comeout(pack *proto.GateOutPack) {
	//log.Println("coming out", pack, "fid2frontend", this.fid2frontend)
	l := len(this.fids)
	if l == 0 {
		return
	}

	switch pack.GetAction() {
	case proto.Action_Broadcast:
		p := this.doPack(pack, this.broadcastFid())
		for _, conn := range this.fid2frontend {
			conn.Send(p)
		}
	case proto.Action_Randomcast:
		rfid := this.randomFid()
		p := this.doPack(pack, rfid)
		if cc := this.fid2frontend[rfid]; cc != nil {
			cc.Send(p)
		} else {
			log.Println("[Error]random not find fid2frontend", rfid)
		}
	case proto.Action_Unicast:
		fid := pack.GetFid()
		if fid == 0 {
			fid = this.uid2fid[pack.GetUid()]
		}
		if fid != 0 {
			if cc := this.fid2frontend[fid]; cc != nil {
				cc.Send(this.doPack(pack, fid))
			} else {
				n_fid := this.randomFid()
				log.Println("[Info]not find fid2frontend", fid, "redirect to", n_fid)
				cc.Send(this.doPack(pack, n_fid))
			}
		} else {
			log.Println("[Error]not find uid2fid", pack.GetUid())
		}
	}
}

func (this *BackGate) doPack(pack *proto.GateOutPack, fid uint32) (ret []byte) {
	fp := &proto.FrontendPack{Uri: pack.Uri, Tsid: pack.Tsid, Ssid: pack.Ssid,
		Bin: pack.Bin, Fid: pb.Uint32(fid)}
	switch pack.GetAction() {
	case proto.Action_Unicast:
		fp.Uid = pack.Uid
	}
	if data, err := pb.Marshal(fp); err == nil {
		uri_field := make([]byte, LEN_URI)
		binary.LittleEndian.PutUint32(uri_field, uint32(URI_TRANSPORT))
		ret = append(uri_field, data...)
	} else {
		log.Println("[Error]pack FrontendPack", err)
	}
	return
}

func (this *BackGate) register(b []byte, cc *ClientConnection) {
	fp := &proto.FrontendRegister{}
	if err := pb.Unmarshal(b, fp); err == nil {
		fid := uint32(fp.GetFid())
		if fid == 0 {
			log.Println("[Error]fid 0 err")
			return
		}
		cc_ex, exist := this.fid2frontend[fid]
		if exist {
			this.unregister(cc_ex)
		}
		this.fid2frontend[fid] = cc
		this.fids = append(this.fids, fid)
		this.notifyRegister(fid)
		log.Println("[Info]register fid:", fid)
	} else {
		log.Println("[Error]Unmarshal register pack", err)
	}
}

func (this *BackGate) unregister(cc *ClientConnection) {
	for fid, c := range this.fid2frontend {
		if c == cc {
			//this.fid2frontend[fid] = nil
			delete(this.fid2frontend, fid)
			log.Println("[Info]unregister", fid)
			for i, f := range this.fids {
				if fid == f {
					last := len(this.fids) - 1
					this.fids[i] = this.fids[last]
					this.fids = this.fids[:last]
				}
			}
		}
	}
}

func (this *BackGate) notifyRegister(fid uint32) {
	if data, err := pb.Marshal(&proto.N2SRegister{Fid: pb.Uint32(fid)}); err == nil {
		this.GateInChan <- &proto.GateInPack{Tsid: pb.Uint32(0), Ssid: pb.Uint32(0),
			Uri: pb.Uint32(99), Bin: data}
	}
}
