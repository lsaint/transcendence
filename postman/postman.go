package postman

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"transcendence/conf"
)

type PostRequest struct {
	Url     string
	Request string
	Ret     chan string
	Sn      int64
}

type Postman struct {
	DoneChan chan *PostRequest
}

func NewPostman() *Postman {
	pm := &Postman{make(chan *PostRequest, conf.CF.BUF_QUEUE)}
	return pm
}

func (this *Postman) post(req *PostRequest) {
	b := strings.NewReader(req.Request)
	//fmt.Println("post-", req.Request, len(req.Request))
	http_req, err := http.Post(req.Url, "application/json", b)
	if err == nil {
		if body, e := ioutil.ReadAll(http_req.Body); e == nil {
			req.Ret <- string(body)
		} else {
			fmt.Println("http post ret err:", e)
		}
		http_req.Body.Close()
	} else {
		close(req.Ret)
	}
}

func (this *Postman) PostAsync(url, s string, sn int64) {
	go func() {
		req := &PostRequest{url, s, make(chan string, 1), sn}
		this.post(req)
		select {
		case this.DoneChan <- req:

		case <-time.After(time.Duration(conf.CF.POST_TIME_OUT) * time.Second):
			fmt.Println("DoneChan timeout", url, sn)
		}
	}()
}

func (this *Postman) Post(url, s string) string {
	req := &PostRequest{url, s, make(chan string, 1), 0}
	go func() {
		this.post(req)
	}()
	select {
	case ret := <-req.Ret:
		return ret

	case <-time.After(time.Duration(conf.CF.POST_TIME_OUT) * time.Second):
		fmt.Println("Post timeout", url)
	}
	return ""
}
