package pymodule

import (
	"container/list"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/qiniu/py"
)

type Task struct {
	Name *py.Base
	Args []*py.Base
}

type TaskMgr struct {
	sync.RWMutex
	task_queue *list.List
	file       *os.File
}

func NewTaskMgr() *TaskMgr {
	file, err := ioutil.TempFile("./", "trans-temp-file-")
	if err != nil {
		log.Fatalln("create temp file err:", err)
	}

	task_queue := list.New()

	mgr := &TaskMgr{file: file, task_queue: task_queue}

	return mgr
}

func (this *TaskMgr) GetFd() int64 {
	return int64(this.file.Fd())
}

func (this *TaskMgr) Notify() {
	this.file.WriteAt([]byte(" "), 0)
}

func (this *TaskMgr) GetTask() []*Task {
	this.Lock()
	defer this.Unlock()
	ret := make([]*Task, 0)
	for e := this.task_queue.Front(); e != nil; e = e.Next() {
		//fmt.Println(e.Value)
		ret = append(ret, e.Value.(*Task))
	}
	this.task_queue.Init()
	return ret
}

func (this *TaskMgr) PushTask(task *Task) {
	this.Lock()
	defer this.Unlock()
	this.task_queue.PushBack(task)
	this.Notify()
}
