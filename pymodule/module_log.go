package pymodule

import (
	"fmt"
	"log"

	"github.com/qiniu/py"
)

var LV = map[string]int{
	"debug": 1,
	"info":  2,
	"warn":  3,
	"error": 4,
	"fatal": 5,
}

var LV_R = map[int]string{
	1: "[DEBUG]",
	2: "[INFO]",
	3: "[WARN]",
	4: "[ERROR]",
	5: "[FATAL]",
}

type LogModule struct {
	lv      int
	logChan chan *LogItem
}

type LogItem struct {
	Lv      int
	Content string
}

func NewLogModule() *LogModule {
	mod := &LogModule{logChan: make(chan *LogItem, 10240)}
	go mod.logging()
	return mod
}

func (this *LogModule) Py_setlevel(args *py.Tuple) (ret *py.Base, err error) {
	var level string
	err = py.Parse(args, &level)
	if err != nil {
		fmt.Println("setlevel err:", err)
		return
	}
	lv, ok := LV[level]
	if !ok {
		fmt.Println("not exist log level", level)
		return
	}
	this.lv = lv
	log.Println("set log level:", level, lv)
	return py.IncNone(), nil
}

func (this *LogModule) logging() {
	for item := range this.logChan {
		log.Println(LV_R[item.Lv], item.Content)
	}
}

func (this *LogModule) DoLog(lv int, args *py.Tuple) {
	var content string
	err := py.Parse(args, &content)
	if err != nil {
		fmt.Println("Log err:", err)
		return
	}

	if this.lv <= lv {
		this.logChan <- &LogItem{Lv: lv, Content: content}
	}
}

func (this *LogModule) Py_debug(args *py.Tuple) (ret *py.Base, err error) {
	this.DoLog(1, args)
	return py.IncNone(), nil
}

func (this *LogModule) Py_info(args *py.Tuple) (ret *py.Base, err error) {
	this.DoLog(2, args)
	return py.IncNone(), nil
}

func (this *LogModule) Py_warn(args *py.Tuple) (ret *py.Base, err error) {
	this.DoLog(3, args)
	return py.IncNone(), nil
}

func (this *LogModule) Py_error(args *py.Tuple) (ret *py.Base, err error) {
	this.DoLog(4, args)
	return py.IncNone(), nil
}

func (this *LogModule) Py_fatal(args *py.Tuple) (ret *py.Base, err error) {
	this.DoLog(5, args)
	return py.IncNone(), nil
}
