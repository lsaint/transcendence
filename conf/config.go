package conf

import (
	"fmt"

	"github.com/qiniu/py"
)

var CF *Config

func init() {
	CF = NewConfig()
}

type Config struct {
	mod *py.Module

	NODE             string
	V1               int
	V2               int
	FID              int
	DIAL_HIVE_ADDR   string
	HIVE_LISTEN_ADDR string
	BUF_QUEUE        int
	MAX_LEN_HEAD     int

	SVCTYPE         int
	SAL_LOCAL_ADDR  string
	SAL_REMOTE_ADDR string

	HTTP_LISTEN_PORT string
	HTTP_LISTEN_URL1 string
	HTTP_LISTEN_URL2 string
	HTTP_TIME_OUT    int

	POST_TIME_OUT int
}

func NewConfig() *Config {
	cf := new(Config)

	code, err := py.CompileFile("./conf/config.py", py.FileInput)
	if err != nil {
		fmt.Println(err)
		panic("Compile config failed")
	}
	defer code.Decref()

	cf.mod, err = py.ExecCodeModule("conf", code.Obj())
	if err != nil {
		fmt.Println(err)
		panic("ExecCodeModule failed")
	}
	defer cf.mod.Decref()

	cf.ReadConfig()

	return cf
}

func (this *Config) ReadConfig() {
	this.NODE = this.getStr("NODE")
	this.V1 = this.getInt("V1")
	this.V2 = this.getInt("V2")
	this.FID = this.getInt("FID")
	this.DIAL_HIVE_ADDR = this.getStr("DIAL_HIVE_ADDR")
	this.HIVE_LISTEN_ADDR = this.getStr("HIVE_LISTEN_ADDR")
	this.BUF_QUEUE = this.getInt("BUF_QUEUE")
	this.MAX_LEN_HEAD = this.getInt("MAX_LEN_HEAD")
	this.SVCTYPE = this.getInt("SVCTYPE")
	this.SAL_LOCAL_ADDR = this.getStr("SAL_LOCAL_ADDR")
	this.SAL_REMOTE_ADDR = this.getStr("SAL_REMOTE_ADDR")
	this.HTTP_LISTEN_PORT = this.getStr("HTTP_LISTEN_PORT")
	this.HTTP_LISTEN_URL1 = this.getStr("HTTP_LISTEN_URL1")
	this.HTTP_LISTEN_URL2 = this.getStr("HTTP_LISTEN_URL2")
	this.HTTP_TIME_OUT = this.getInt("HTTP_TIME_OUT")
	this.POST_TIME_OUT = this.getInt("POST_TIME_OUT")
}

func (this *Config) getStr(attr string) string {
	s, err := this.mod.GetAttrString(attr)
	defer s.Decref()
	if err != nil {
		fmt.Println(err)
		panic("Config getStr err")
	}
	ss, ok := py.ToString(s)
	if !ok {
		panic("Config ToString err")
	}
	return ss
}

func (this *Config) getInt(attr string) int {
	s, err := this.mod.GetAttrString(attr)
	defer s.Decref()
	if err != nil {
		fmt.Println(err)
		panic("Config getInt err")
	}
	i, ok := py.ToInt(s)
	if !ok {
		panic("Config ToInt err")
	}
	return i
}
