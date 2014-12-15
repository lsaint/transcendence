package network

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	. "transcendence/conf"
)

type RequestCtx struct {
	Url  string
	Body string
	Ret  chan string
	Sn   int64
}

type Postman struct {
	DoneChan chan *RequestCtx
	client   *http.Client
}

func NewPostman() *Postman {
	client := &http.Client{Timeout: time.Duration(I("POST_TIME_OUT")) * time.Second}
	pm := &Postman{make(chan *RequestCtx, I("BUF_QUEUE")), client}
	return pm
}

func (this *Postman) post(req *RequestCtx) {
	b := strings.NewReader(req.Body)
	resp, err := this.client.Post(req.Url, "application/json", b)
	if err == nil {
		if body, e := ioutil.ReadAll(resp.Body); e == nil {
			req.Ret <- string(body)
		} else {
			fmt.Println("http post ret err:", e)
		}
		resp.Body.Close()
	} else {
		close(req.Ret)
	}
}

func (this *Postman) PostAsync(url, s string, sn int64) {
	go func() {
		req := &RequestCtx{url, s, make(chan string, 1), sn}
		this.post(req)
		this.DoneChan <- req
	}()
}

func (this *Postman) Post(url, s string) string {
	req := &RequestCtx{url, s, make(chan string, 1), 0}
	this.post(req)
	return <-req.Ret
}
