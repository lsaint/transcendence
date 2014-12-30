package conf

import (
	"flag"
	"log"

	"github.com/lsaint/py"
)

var config *Config

func init() {
	config = NewConfig()
}

//

func S(n string) string {
	return config.getStr(n)
}

func I(n string) int {
	return config.getInt(n)
}

func TS(n string) []string {
	return config.getTupleStr(n)
}

//

type Config struct {
	mod *py.Module
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
		log.Fatalln("Exec conf.py failed", err)
	}
	defer cf.mod.Decref()

	return cf
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
