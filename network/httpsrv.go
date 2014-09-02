package network

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"transcendence/conf"
)

type HttpReq struct {
	Req string
	Ret chan string
	Url string
}

type HttpServer struct {
	reqChan chan *HttpReq
}

func NewHttpServer(c chan *HttpReq) *HttpServer {
	return &HttpServer{c}
}

func (this *HttpServer) Start() {
	http.HandleFunc(conf.CF.HTTP_LISTEN_URL1, func(w http.ResponseWriter, r *http.Request) {
		this.onReq(w, r, conf.CF.HTTP_LISTEN_URL1)
	})
	http.HandleFunc(conf.CF.HTTP_LISTEN_URL2, func(w http.ResponseWriter, r *http.Request) {
		this.onReq(w, r, conf.CF.HTTP_LISTEN_URL2)
	})

	log.Println("http server running, listen ",
		conf.CF.HTTP_LISTEN_PORT, conf.CF.HTTP_LISTEN_URL1, conf.CF.HTTP_LISTEN_URL2)
	http.ListenAndServe(conf.CF.HTTP_LISTEN_PORT, nil)
}

func (this *HttpServer) onReq(w http.ResponseWriter, r *http.Request, url string) {
	recv_post, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprint(w, "")
		return
	}
	ret, ret_chan := "", make(chan string)
	select {
	case this.reqChan <- &HttpReq{string(recv_post), ret_chan, url}:
		ret = <-ret_chan

	case <-time.After(time.Duration(conf.CF.HTTP_TIME_OUT) * time.Second):
	}

	fmt.Fprint(w, ret)
}
