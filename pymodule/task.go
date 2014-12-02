package pymodule

//#ifdef __linux__
//#include <sys/eventfd.h>
//#else
//int eventfd(unsigned int initval, int flags){return 0;}
//#endif
import "C"
import (
	"container/list"
	"encoding/binary"
	"os"
	"runtime"
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
	fd         int64
	buf        []byte
}

func NewTaskMgr() *TaskMgr {
	task_queue := list.New()

	mgr := &TaskMgr{task_queue: task_queue,
		buf: make([]byte, 8),
		fd:  int64(C.eventfd(0, 0))}

	return mgr
}

func (this *TaskMgr) GetFd() int64 {
	return this.fd
}

func (this *TaskMgr) Notify() {
	if runtime.GOOS == "linux" {
		binary.PutUvarint(this.buf, 1)
		syscall.Write(int(this.fd), this.buf)
	} else {
		syscall.Kill(os.Getpid(), syscall.SIGUSR1)
	}
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
