package conf

import (
	"flag"
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
	HTTP_LISTEN_URLS []string
	HTTP_TIME_OUT    int

	POST_TIME_OUT int

	CLUSTER_NODE_NAME     string
	CLUSTER_NODE_PORT     int
	CLUSTER_NODE_CONNECT2 string

	RAFT_ADDR string
	RAFT_DIR  string

	HOST_SERVICE        string
	URL_SERVICE_REG     string
	SERVICE_LISTEN_PORT int
	CTX_REG             string
}

func NewConfig() *Config {
	conf_path := flag.String("c", "./conf/config.py", "config file path")
	flag.Parse()

	cf := new(Config)
	code, err := py.CompileFile(*conf_path, py.FileInput)
	if err != nil {
		log.Fatalln("Compile config failed", err)
	}
	defer code.Decref()

	cf.mod, err = py.ExecCodeModule("conf", code.Obj())
	if err != nil {
		log.Fatalln("ExecCodeModule failed", err)
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
	this.HTTP_LISTEN_URLS = this.getTupleStr("HTTP_LISTEN_URLS")
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
		log.Fatalln("Config getStr err", err)
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
		log.Fatalln("Config getInt err", err)
	}
	i, ok := py.ToInt(s)
	if !ok {
		log.Fatalln("Config ToInt err")
	}
	return i
}

func (this *Config) getTuple(attr string) *py.Tuple {
	t, err := this.mod.GetAttrString(attr)
	if err != nil {
		log.Fatalln("Config getTuple err", err)
	}
	tp, ok := py.AsTuple(t)
	if !ok {
		log.Fatalln("Config AsTuple err")
	}
	return tp
}

func (this *Config) getTupleStr(attr string) []string {
	tp := this.getTuple(attr)
	defer tp.Decref()

	slice := tp.Slice()
	ret := make([]string, 0)
	for _, obj := range slice {
		s, ok := py.ToString(obj)
		if !ok {
			log.Fatalln("getTupleStr to str err")
		}
		ret = append(ret, s)
	}
	return ret
}
