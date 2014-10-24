package network

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
	"transcendence/conf"

	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-mdb"
	"github.com/ugorji/go/codec"
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

	raftNode *RaftNode
}

func NewClusterNode() *ClusterNode {
	ch := make(chan NodeEvent, 1024)
	node := &ClusterNode{NodeEventChan: ch}

	// raft
	node.raftNode = NewRaftNode()
	go node.waitLeader(node.raftNode.LeaderCh)

	// memberlist
	config := memberlist.DefaultLocalConfig()
	config.Name = conf.CF.CLUSTER_NODE_NAME
	config.BindAddr = "127.0.0.1"
	config.BindPort = conf.CF.CLUSTER_NODE_PORT
	config.Events = node
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
	log.Println("[CLUSTER]raft adding addr", addr)
	c.NodeEventChan <- NodeEvent{NodeJoin, n}
	if c.raftNode.IsLeader() {
		future := c.raftNode.Raft.AddPeer(addr)
		if err := future.Error(); err != nil {
			log.Println("[CLUSTER] addpeer fail", addr, err)
		}
	}
}

func (c *ClusterNode) NotifyLeave(n *memberlist.Node) {
	addr := &net.TCPAddr{IP: n.Addr, Port: int(n.Port) + 100}
	log.Println("[CLUSTER]raft removing addr", addr)
	c.NodeEventChan <- NodeEvent{NodeLeave, n}
	//if c.raftNode.IsLeader() {
	//	future := c.raftNode.Raft.RemovePeer(addr)
	//	if err := future.Error(); err != nil {
	//		log.Println("[CLUSTER] remove peer fail", addr, err)
	//	}
	//}
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

func (c *ClusterNode) waitLeader(ch <-chan bool) {
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

type RaftNode struct {
	trans *raft.NetworkTransport
	peers *raft.JSONPeers

	Raft     *raft.Raft
	LeaderCh <-chan bool
}

func NewRaftNode() *RaftNode {
	var err error
	node := &RaftNode{}

	config := raft.DefaultConfig()
	config.EnableSingleNode = true

	node.trans, err = raft.NewTCPTransport(conf.CF.RAFT_ADDR,
		nil, 2, time.Second, config.LogOutput)
	if err != nil {
		log.Fatalln("trans err", err)
	}

	path := conf.CF.RAFT_DIR
	dbSize := uint64(8 * 1024 * 1024)
	store, err := raftmdb.NewMDBStoreWithSize(path, dbSize)
	if err != nil {
		log.Fatalln("store err", err)
	}

	snapshotsRetained := 2
	snapshots, err := raft.NewFileSnapshotStore(path, snapshotsRetained, config.LogOutput)
	if err != nil {
		log.Fatalln("snapshots err", err)
	}

	node.peers = raft.NewJSONPeers(path, node.trans)

	fsm := &MockFSM{}

	node.Raft, err = raft.NewRaft(config, fsm, store, store, snapshots, node.peers, node.trans)
	if err != nil {
		log.Fatalln("make raft err:", err)
	}

	node.LeaderCh = node.Raft.LeaderCh()
	return node
}

func (this *RaftNode) IsLeader() bool {
	return this.Raft.Leader() == this.trans.LocalAddr()
}

func (this *RaftNode) Peers() ([]net.Addr, error) {
	return this.peers.Peers()
}

// MockFSM is an implementation of the FSM interface, and just stores
// the logs sequentially
type MockFSM struct {
	sync.Mutex
	logs [][]byte
}

type MockSnapshot struct {
	logs     [][]byte
	maxIndex int
}

func (m *MockFSM) Apply(log *raft.Log) interface{} {
	m.Lock()
	defer m.Unlock()
	m.logs = append(m.logs, log.Data)
	return len(m.logs)
}

func (m *MockFSM) Snapshot() (raft.FSMSnapshot, error) {
	m.Lock()
	defer m.Unlock()
	return &MockSnapshot{m.logs, len(m.logs)}, nil
}

func (m *MockFSM) Restore(inp io.ReadCloser) error {
	m.Lock()
	defer m.Unlock()
	defer inp.Close()
	hd := codec.MsgpackHandle{}
	dec := codec.NewDecoder(inp, &hd)

	m.logs = nil
	return dec.Decode(&m.logs)
}

func (m *MockSnapshot) Persist(sink raft.SnapshotSink) error {
	hd := codec.MsgpackHandle{}
	enc := codec.NewEncoder(sink, &hd)
	if err := enc.Encode(m.logs[:m.maxIndex]); err != nil {
		sink.Cancel()
		return err
	}
	sink.Close()
	return nil
}

func (m *MockSnapshot) Release() {
}

func (m MockSnapshot) Write(p []byte) (n int, err error) {
	return
}
