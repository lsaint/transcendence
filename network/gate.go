package network

import (
	"encoding/binary"
	"log"
	"net"

	pb "code.google.com/p/goprotobuf/proto"

	. "transcendence/conf"
	"transcendence/proto"
)

const ()

type GateServer struct {
	lidCounter int64
	buffChan   chan *ConnBuff
	lid2conn   map[int64]*IConnection
	lid2sid    map[int64]uint32
	sid2conns  map[uint32][]*IConnection
	GateEntry  chan *proto.GateInPack
	GateExit   chan *proto.GateOutPack
}

func NewGateServer(entry chan *proto.GateInPack, exit chan *proto.GateOutPack) *GateServer {
	gs := &GateServer{buffChan: make(chan *ConnBuff, I("BUF_QUEUE")),
		lidCounter: 0,
		lid2conn:   make(map[int64]*IConnection),
		lid2sid:    make(map[int64]uint32),
		sid2conns:  make(map[uint32][]*IConnection),
		GateEntry:  entry,
		GateExit:   exit}
	go gs.parse()
	return gs
}

func (this *GateServer) Start() {
	ln, err := net.Listen("tcp", S("HIVE_LISTEN_ADDR"))
	if err != nil {
		log.Fatalln("Listen err", err)
	}
	log.Println("[Info]BackGate running", S("HIVE_LISTEN_ADDR"))
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("[Error]Accept error", err)
			continue
		}
		this.lidCounter += 1
		go this.acceptConn(conn, this.lidCounter)
	}
}

func (this *GateServer) acceptConn(conn net.Conn, lid int64) {
	cliConn := NewIConnection(conn)
	for {
		if buff_body, ok := cliConn.ReadBody(); ok {
			this.buffChan <- &ConnBuff{cliConn, buff_body, lid}
			continue
		}
		this.buffChan <- &ConnBuff{cliConn, nil, lid}
		break
	}
	cliConn.Close()
}

func (this *GateServer) parse() {
	go this.comeout()

	for conn_buff := range this.buffChan {
		msg := conn_buff.buff
		conn := conn_buff.conn
		lid := conn_buff.lid

		if msg == nil {
			this.unregister(conn, lid, false)
			continue
		}
		if len(msg) < LEN_URI {
			continue
		}

		uri := binary.LittleEndian.Uint16(msg[:LEN_URI])
		sid := uint32(0)
		if uri == 1 {
			sid = this.get_sid_when_uri1()
		} else if _, exist := this.lid2sid[lid]; exist {
			sid = this.lid2sid[lid]
		} else {
			log.Println("[Error]sid not exist")
			continue
		}
		if _, exist := this.lid2conn[lid]; !exist {
			this.register(conn, lid, sid)
		}
		this.comein(msg[LEN_URI:])

	}
}

func (this *GateServer) comeout() {
	for gop := range this.GateExit {
		lids := gop.GetLids()
		if len(lids) != 0 {
			for _, lid := range lids {
				if conn, ok := this.lid2conn[lid]; ok {
					if gop.GetUri() == 0 {
						this.unregister(conn, lid, true)
					} else {
						conn.Send(gop.GetBin())
					}
				}
			}
		} else {
			if sid := gop.GetSid(); sid != 0 {
				this.Broadcast(sid, gop.GetBin())
			}
		}
	}
}

// test
func (this *GateServer) get_sid_when_uri1() uint32 {
	return 0
}

func (this *GateServer) unpack(b []byte) (msg *proto.GateInPack, err error) {
	return
}

func (this *GateServer) comein(b []byte) {
	if msg, err := this.unpack(b); err == nil {
		this.GateEntry <- msg
	}
}

func (this *GateServer) register(conn *IConnection, lid int64, sid uint32) {
	this.lid2conn[lid] = conn
	this.lid2sid[lid] = sid
	if conns, exist := this.sid2conns[sid]; exist {
		this.sid2conns[sid] = append(conns, conn)
	} else {
		conns := make([]*IConnection, 0, 16)
		conns = append(conns, conn)
		this.sid2conns[sid] = conns
	}
}

func (this *GateServer) unregister(conn *IConnection, lid int64, initiative bool) {
	sid := this.lid2sid[lid]
	delete(this.lid2sid, lid)
	conns := this.sid2conns[sid]
	for i, c := range conns {
		if c == conn {
			conns[i] = conns[len(conns)-1]
			conns = conns[:len(conns)-1]
		}
	}
	this.sid2conns[sid] = conns

	if c, exist := this.lid2conn[lid]; exist {
		delete(this.lid2conn, lid)
		if initiative {
			c.Close()
		} else {
			this.GateEntry <- &proto.GateInPack{Lid: pb.Int64(lid), Sid: pb.Uint32(sid)}
		}
	}
}

func (this *GateServer) Broadcast(sid uint32, bin []byte) {
	if conns, ok := this.sid2conns[sid]; ok {
		for _, conn := range conns {
			conn.Send(bin)
		}
	}
}

type ConnBuff struct {
	conn *IConnection
	buff []byte
	lid  int64
}
