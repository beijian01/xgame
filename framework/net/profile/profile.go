package profile

import (
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	NodeTypeGate   = "gate"
	NodeTypeAuth   = "auth"
	NodeTypeGame   = "game"
	NodeTypeEntity = "entity"
)

type ClusterCfg struct {
	Project string           `json:"project"`
	Etcd    *clientv3.Config `json:"etcd"`
	Nodes   []*NodeCfg       `json:"nodes"`
	Nats    *NatsCfg         `json:"nats"`
	//Redis *RedisCfg        `json:"redis"`
	//Mysql *MysqlCfg        `json:"mysql"`
}

func (c *ClusterCfg) FindNode(nodeId string) (*NodeCfg, bool) {
	for _, node := range c.Nodes {
		if node.NodeId == nodeId {
			return node, true
		}
	}
	return nil, false
}

type NodeCfg struct {
	NodeId   string         `json:"node_id"`
	NodeType string         `json:"node_type"`
	Ports    map[string]int `json:"ports,omitempty"`
	IsGate   bool           `json:"is_gate,omitempty"`
}

func (n *NodeCfg) GetListenPorts() map[string]int {
	return n.Ports
}

func (n *NodeCfg) GetNodeId() string {
	return n.NodeId
}

func (n *NodeCfg) GetNodeType() string {
	return n.NodeType
}

type NatsCfg struct {
	Address        string `json:"address"`
	User           string `json:"user,omitempty"`
	Password       string `json:"password,omitempty"`
	ReconnectDelay int    `json:"reconnect_delay,omitempty"`
}
