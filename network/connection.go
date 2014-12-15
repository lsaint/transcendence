package network

import (
	"encoding/binary"
	"io"
	"log"

	. "transcendence/conf"
)

const (
	LEN_HEAD  = 4
	LEN_URI   = 4
	LEN_EXTRA = LEN_HEAD // 总长度包括包长
)

type IConnection struct {
	conn     io.ReadWriteCloser
	sendchan chan []byte
}

func NewIConnection(c io.ReadWriteCloser) *IConnection {
	cliConn := new(IConnection)
	cliConn.conn = c
	cliConn.sendchan = make(chan []byte, I("BUF_QUEUE"))
	go cliConn.sending()
	return cliConn
}

func (this *IConnection) Send(buf []byte) {
	head := make([]byte, LEN_HEAD)
	binary.LittleEndian.PutUint32(head, uint32(len(buf)+LEN_EXTRA))
	buf = append(head, buf...)

	select {
	case this.sendchan <- buf:

	default:
		log.Println("[Error]sendchan overflow or closed")
	}
}

func (this *IConnection) sending() {
	for b := range this.sendchan {
		if _, err := this.conn.Write(b); err != nil {
			log.Println("[Error]conn write err:", err)
			this.Close()
		}
	}
}

func (this *IConnection) Read(buff []byte) bool {
	if _, err := io.ReadFull(this.conn, buff); err != nil {
		log.Println("[Error]ReadFull err", err)
		return false
	}
	return true
}

func (this *IConnection) ReadBody() (ret []byte, ok bool) {
	buff_head := make([]byte, LEN_HEAD)
	if !this.Read(buff_head) {
		return
	}
	len_head := binary.LittleEndian.Uint32(buff_head) - LEN_EXTRA
	if len_head > uint32(I("MAX_LEN_HEAD")) {
		log.Println("[Error]message len too long", len_head, string(buff_head))
		return
	}
	ret = make([]byte, len_head)
	if !this.Read(ret) {
		return
	}
	ok = true
	return
}

func (this *IConnection) Close() {
	this.conn.Close()
	//close(this.sendchan)
}

type ConnBuff struct {
	conn *IConnection
	buff []byte
}
