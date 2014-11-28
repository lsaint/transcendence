package pymodule

import (
	"container/list"
	"os"
	"sync"
	"syscall"

	"github.com/qiniu/py"
)

type Task struct {
	Name *py.Base
	Args []*py.Base
}

type TaskMgr struct {
	sync.RWMutex
	task_queue *list.List
}

func NewTaskMgr() *TaskMgr {
	task_queue := list.New()

	mgr := &TaskMgr{task_queue: task_queue}

	return mgr
}

func (this *TaskMgr) GetFd() int64 {
	return 0
}

func (this *TaskMgr) Notify() {
	syscall.Kill(os.Getpid(), syscall.SIGUSR1)
}

func (this *TaskMgr) GetTask() []*Task {
	this.Lock()
	defer this.Unlock()
	ret := make([]*Task, 0)
	for e := this.task_queue.Front(); e != nil; e = e.Next() {
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
