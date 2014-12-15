package network

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	. "transcendence/conf"
)

type HttpReq struct {
	Req string
	Ret chan string
	Url string
}

type HttpServer struct {
	reqChan chan *HttpReq
	addr    string
	urls    []string
}

func NewHttpServer(c chan *HttpReq, addr string, urls []string) *HttpServer {
	return &HttpServer{c, addr, urls}
}

func (this *HttpServer) Start() {
	for _, url := range this.urls {
		http.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
			this.onReq(w, r, url)
		})
	}

	log.Println("http server running, listen ", this.addr, this.urls)
	log.Fatalln(http.ListenAndServe(this.addr, nil))
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

	case <-time.After(time.Duration(I("HTTP_TIME_OUT")) * time.Second):
	}

	fmt.Fprint(w, ret)
}
