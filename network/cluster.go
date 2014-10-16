package network

import (
	"log"
	"transcendence/conf"

	. "github.com/hashicorp/memberlist"
)

type ClusterNode struct {
	NodeEventChan chan NodeEvent
}

func NewClusterNode() *ClusterNode {
	ch := make(chan NodeEvent, 1024)
	node := &ClusterNode{NodeEventChan: ch}

	config := DefaultLocalConfig()
	config.Name = conf.CF.CLUSTER_NODE_NAME
	config.BindPort = conf.CF.CLUSTER_NODE_PORT
	config.Events = node
	l, err := Create(config) // memberlist.Create
	if err != nil {
		log.Fatalln("Failed to create memberlist: " + err.Error())
	}

	if conf.CF.CLUSTER_NODE_CONNECT2 != "" {
		_, err = l.Join([]string{conf.CF.CLUSTER_NODE_CONNECT2})
		if err != nil {
			log.Fatalln("Failed to join cluster: " + err.Error())
		}
	}

	return node
}

func (c *ClusterNode) NotifyJoin(n *Node) {
	c.NodeEventChan <- NodeEvent{NodeJoin, n}
}

func (c *ClusterNode) NotifyLeave(n *Node) {
	c.NodeEventChan <- NodeEvent{NodeLeave, n}
}

func (c *ClusterNode) NotifyUpdate(n *Node) {
	c.NodeEventChan <- NodeEvent{NodeUpdate, n}
}
