package pymodule

import (
	"log"

	"github.com/garyburd/redigo/redis"
	"github.com/qiniu/py"
)

type RedisModule struct {
	conn    redis.Conn
	cmdChan chan *RedisCmd
}

type RedisCmd struct {
	Cmd  string
	Args []interface{}
}

func NewRedisModule() *RedisModule {
	mod := &RedisModule{cmdChan: make(chan *RedisCmd, 10240)}
	return mod
}

func (this *RedisModule) doing() {
	for rc := range this.cmdChan {
		this.conn.Do(rc.Cmd, rc.Args...)
	}
}

func (this *RedisModule) Py_dial(args *py.Tuple) (ret *py.Base, err error) {
	var network, address string
	err = py.Parse(args, &network, &address)
	if err != nil {
		log.Fatalln("parse redis addr err", err)
	}
	this.conn, err = redis.Dial(network, address)
	if err != nil {
		log.Fatalln("connect to redis err", err)
	}
	log.Println("dial to redis (", network, address, ") sucess")
	go this.doing()
	return py.IncNone(), nil
}

func (this *RedisModule) Py_do(args *py.Tuple) (ret *py.Base, err error) {
	var cmd string
	var ss []interface{}
	err = py.ParseV(args, &cmd, &ss)
	if err != nil {
		return
	}
	rc := &RedisCmd{Cmd: cmd, Args: ss}
	this.cmdChan <- rc
	return py.IncNone(), nil
}
