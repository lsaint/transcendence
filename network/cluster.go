package network

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
	"transcendence/conf"

	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-mdb"
	"github.com/hashicorp/go-msgpack/codec"
)

type NodeEventType int

const (
	NodeJoin NodeEventType = iota
	NodeLeave
	NodeUpdate
	NodeBecomeLeader
	NodeHandoffLeader
)

type NodeEvent struct {
	Event NodeEventType
	*memberlist.Node
}

type ClusterNode struct {
	NodeEventChan chan NodeEvent

	RaftAgent *ClusterRaftAgent
}

func NewClusterNode() *ClusterNode {
	ch := make(chan NodeEvent, 1024)
	node := &ClusterNode{NodeEventChan: ch}
	applyCh := make(chan *raft.Log, 1024)

	// raft
	node.RaftAgent = NewClusterRaftAgent(applyCh)
	go node.notifyLeader(node.RaftAgent.LeaderCh)

	// memberlist
	config := memberlist.DefaultLocalConfig()
	config.Name = conf.CF.CLUSTER_NODE_NAME
	config.BindAddr = "127.0.0.1"
	config.BindPort = conf.CF.CLUSTER_NODE_PORT
	config.Events = node
	file, err := os.Create(fmt.Sprintf("%v/memlog", conf.CF.RAFT_DIR))
	if err != nil {
		log.Fatalln("Create memlog err:", err)
	}
	config.LogOutput = file
	l, err := memberlist.Create(config) // memberlist.Create
	if err != nil {
		log.Fatalln("Failed to create memberlist: " + err.Error())
	}
	log.Println("[CLUSTER]memberlist addr:", config.BindAddr, config.BindPort)

	if conf.CF.CLUSTER_NODE_CONNECT2 != "" {
		_, err = l.Join([]string{conf.CF.CLUSTER_NODE_CONNECT2})
		if err != nil {
			log.Fatalln("Failed to join cluster: " + err.Error())
		}
	}

	return node
}

func (c *ClusterNode) NotifyJoin(n *memberlist.Node) {
	addr := &net.TCPAddr{IP: n.Addr, Port: int(n.Port) + 100}
	c.NodeEventChan <- NodeEvent{NodeJoin, n}
	if c.RaftAgent.IsLeader() {
		future := c.RaftAgent.Raft.AddPeer(addr)
		if err := future.Error(); err != nil {
			log.Println("[CLUSTER] addpeer fail", addr, err)
		}
	}
}

func (c *ClusterNode) NotifyLeave(n *memberlist.Node) {
	c.NodeEventChan <- NodeEvent{NodeLeave, n}
}

func (c *ClusterNode) NotifyUpdate(n *memberlist.Node) {
	c.NodeEventChan <- NodeEvent{NodeUpdate, n}
}

func (c *ClusterNode) NotifyBecomeLeader() {
	c.NodeEventChan <- NodeEvent{NodeBecomeLeader, nil}
}

func (c *ClusterNode) NotifyHandoffLeader() {
	c.NodeEventChan <- NodeEvent{NodeHandoffLeader, nil}
}

func (c *ClusterNode) notifyLeader(ch <-chan bool) {
	for {
		select {
		case b := <-ch:
			if b {
				c.NotifyBecomeLeader()
			} else {
				c.NotifyHandoffLeader()
			}
		}
	}
}

//

type ClusterRaftAgent struct {
	trans *raft.NetworkTransport
	peers *raft.JSONPeers

	Raft     *raft.Raft
	LeaderCh <-chan bool
	ApplyCh  chan *raft.Log
}

func NewClusterRaftAgent(applyCh chan *raft.Log) *ClusterRaftAgent {
	var err error
	node := &ClusterRaftAgent{ApplyCh: applyCh}
	path := conf.CF.RAFT_DIR

	dbSize := uint64(8 * 1024 * 1024)
	store, err := raftmdb.NewMDBStoreWithSize(path, dbSize)
	if err != nil {
		log.Fatalln("store err", err)
	}

	config := raft.DefaultConfig()
	config.EnableSingleNode = true
	config.SnapshotInterval = 30 * time.Second
	config.SnapshotThreshold = 1
	file, err := os.Create(fmt.Sprintf("%vlogoutput", path))
	if err != nil {
		log.Fatalln("create logoutput err:", err)
	}
	config.LogOutput = file

	node.trans, err = raft.NewTCPTransport(conf.CF.RAFT_ADDR,
		nil, 2, time.Second, config.LogOutput)
	if err != nil {
		log.Fatalln("trans err", err)
	}

	snapshotsRetained := 2
	snapshots, err := raft.NewFileSnapshotStore(path, snapshotsRetained, config.LogOutput)
	if err != nil {
		log.Fatalln("snapshots err", err)
	}

	node.peers = raft.NewJSONPeers(path, node.trans)

	fsm := &FSM{ApplyCh: applyCh}

	node.Raft, err = raft.NewRaft(config, fsm, store, store, snapshots, node.peers, node.trans)
	if err != nil {
		log.Fatalln("make raft err:", err)
	}

	node.LeaderCh = node.Raft.LeaderCh()
	return node
}

func (this *ClusterRaftAgent) IsLeader() bool {
	return this.Raft.Leader() == this.trans.LocalAddr()
}

func (this *ClusterRaftAgent) Peers() ([]net.Addr, error) {
	return this.peers.Peers()
}

type FSM struct {
	sync.Mutex
	logs    [][]byte
	ApplyCh chan *raft.Log
}

type FSMSnapshot struct {
	logs     [][]byte
	maxIndex int
}

func (m *FSM) Apply(log *raft.Log) interface{} {
	//fmt.Println("FSM.Apply.Data:", string(log.Data))
	m.ApplyCh <- log
	m.Lock()
	defer m.Unlock()
	m.logs = append(m.logs, log.Data)
	return len(m.logs)
}

func (m *FSM) Snapshot() (raft.FSMSnapshot, error) {
	fmt.Println("FSM.Snapshoting")
	m.Lock()
	defer m.Unlock()
	return &FSMSnapshot{m.logs, len(m.logs)}, nil
}

func (m *FSM) Restore(inp io.ReadCloser) error {
	fmt.Println("###FSM.Restore###")
	m.Lock()
	defer m.Unlock()
	defer inp.Close()
	hd := codec.MsgpackHandle{}
	dec := codec.NewDecoder(inp, &hd)

	m.logs = nil
	return dec.Decode(&m.logs)
}

func (m *FSMSnapshot) Persist(sink raft.SnapshotSink) error {
	fmt.Println("FSMSnapshot.Persist")
	hd := codec.MsgpackHandle{}
	enc := codec.NewEncoder(sink, &hd)
	if err := enc.Encode(m.logs[:m.maxIndex]); err != nil {
		sink.Cancel()
		return err
	}
	sink.Close()
	return nil
}

func (m *FSMSnapshot) Release() {
}

func (m FSMSnapshot) Write(p []byte) (n int, err error) {
	return
}
