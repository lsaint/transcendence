package pymodule

import (
	"errors"
	"log"
	"time"
	"transcendence/network"

	"github.com/lsaint/py"
)

type RaftModule struct {
	cn *network.ClusterNode
}

func NewRaftModule(cn *network.ClusterNode) *RaftModule {
	mod := &RaftModule{cn: cn}
	return mod
}

func (this *RaftModule) Py_apply(args *py.Tuple) (ret *py.Base, err error) {
	if !this.cn.RaftAgent.IsLeader() {
		return py.IncNone(), errors.New("not leader")
	}

	var cmd string
	err = py.Parse(args, &cmd)
	if err != nil {
		log.Println("py raft apply err:", err)
		return
	}
	go this.cn.RaftAgent.Raft.Apply([]byte(cmd), 30*time.Second)
	return py.IncNone(), nil
}

func (this *RaftModule) Py_leader(args *py.Tuple) (ret *py.Base, err error) {
	addr := this.cn.RaftAgent.Raft.Leader().String()
	return py.NewString(addr).Obj(), nil
}
