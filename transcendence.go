// L'transcendence

package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	. "transcendence/conf"
	"transcendence/network"
	"transcendence/proto"
	"transcendence/pymodule"
)

func handleSig() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	for sig := range c {
		log.Println("ABORT BY", sig)
		return
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	//go func() {
	//    log.Println(http.ListenAndServe("localhost:6061", nil))
	//}()

	in := make(chan *proto.Passpack, CF.BUF_QUEUE)
	out := make(chan *proto.Passpack, CF.BUF_QUEUE)
	http_req_chan := make(chan *network.HttpReq, CF.BUF_QUEUE)
	pymgr := pymodule.NewPyMgr(in, out, http_req_chan)
	httpsrv := network.NewHttpServer(http_req_chan, CF.HTTP_LISTEN_ADDR, CF.HTTP_LISTEN_URLS)

	go pymgr.Start()
	go httpsrv.Start()

	handleSig()
}
