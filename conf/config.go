package conf

import (
	"flag"
	"fmt"
	"log"

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

	CLUSTER_NODE_NAME     string
	CLUSTER_NODE_PORT     int
	CLUSTER_NODE_CONNECT2 string

	RAFT_ADDR string
	RAFT_DIR  string
}

func NewConfig() *Config {
	conf_path := flag.String("c", "./conf/config.py", "config file path")
	flag.Parse()

	cf := new(Config)
	code, err := py.CompileFile(*conf_path, py.FileInput)
	if err != nil {
		fmt.Println(err)
		log.Fatalln("Compile config failed")
	}
	defer code.Decref()

	cf.mod, err = py.ExecCodeModule("conf", code.Obj())
	if err != nil {
		fmt.Println(err)
		log.Fatalln("ExecCodeModule failed")
	}
	defer cf.mod.Decref()

	cf.ReadConfig()

	return cf
}
func (this *Config) ReadConfig() {
	this.V1 = this.getInt("V1")
	this.V2 = this.getInt("V2")
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
	this.CLUSTER_NODE_NAME = this.getStr("CLUSTER_NODE_NAME")
	this.CLUSTER_NODE_PORT = this.getInt("CLUSTER_NODE_PORT")
	this.CLUSTER_NODE_CONNECT2 = this.getStr("CLUSTER_NODE_CONNECT2")
	this.RAFT_ADDR = this.getStr("RAFT_ADDR")
	this.RAFT_DIR = this.getStr("RAFT_DIR")
}

func (this *Config) getStr(attr string) string {
	s, err := this.mod.GetAttrString(attr)
	defer s.Decref()
	if err != nil {
		fmt.Println(err)
		log.Fatalln("Config getStr err")
	}
	ss, ok := py.ToString(s)
	if !ok {
		log.Fatalln("Config ToString err")
	}
	return ss
}

func (this *Config) getInt(attr string) int {
	s, err := this.mod.GetAttrString(attr)
	defer s.Decref()
	if err != nil {
		fmt.Println(err)
		log.Fatalln("Config getInt err")
	}
	i, ok := py.ToInt(s)
	if !ok {
		log.Fatalln("Config ToInt err")
	}
	return i
}
