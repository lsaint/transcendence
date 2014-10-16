// L'transcendence

package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"transcendence/conf"
	"transcendence/network"
	"transcendence/proto"
	"transcendence/pymodule"
)

func handleSig() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	for sig := range c {
		log.Println("__handle__signal__", sig)
		return
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	//go func() {
	//    log.Println(http.ListenAndServe("localhost:6061", nil))
	//}()

	cluster_node := network.NewClusterNode()
	in := make(chan *proto.Passpack, conf.CF.BUF_QUEUE)
	out := make(chan *proto.Passpack, conf.CF.BUF_QUEUE)
	http_req_chan := make(chan *network.HttpReq, conf.CF.BUF_QUEUE)
	pymgr := pymodule.NewPyMgr(in, out, http_req_chan, cluster_node.NodeEventChan)
	httpsrv := network.NewHttpServer(http_req_chan)

	if conf.CF.NODE == "hive" {
		go network.NewBackGate(in, out).Start()
	} else if conf.CF.NODE == "drone" {
		go network.NewFrontGate(in, out).Start()
	} else {
		log.Fatalln("invalid node")
	}

	go pymgr.Start()
	go httpsrv.Start()

	handleSig()
}
