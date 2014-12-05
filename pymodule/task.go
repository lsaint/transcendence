package pymodule

//#ifdef __linux__
//#include <sys/eventfd.h>
//#else
//int eventfd(unsigned int initval, int flags){return 0;}
//#endif
import "C"
import (
	"encoding/binary"

	"os"
	"runtime"
	"syscall"
)

type TaskWaitress struct {
	fd      int64
	buf     []byte
	waiting chan bool
	pyready chan bool
}

func NewTaskWaitress(waiting, pyready chan bool) *TaskWaitress {
	buf := make([]byte, 8)
	binary.PutUvarint(buf, 1)

	mgr := &TaskWaitress{
		buf:     buf,
		waiting: waiting,
		pyready: pyready,
		fd:      int64(C.eventfd(0, 0))}
	return mgr
}

func (this *TaskWaitress) GetFd() int64 {
	return this.fd
}

func (this *TaskWaitress) Notify() {
	if runtime.GOOS == "linux" {
		syscall.Write(int(this.fd), this.buf)
	} else {
		syscall.Kill(os.Getpid(), syscall.SIGUSR1)
	}
}

func (this *TaskWaitress) PyReady() {
	this.pyready <- true
}

func (this *TaskWaitress) DoTask() {
	this.waiting <- true
	<-this.waiting
}
