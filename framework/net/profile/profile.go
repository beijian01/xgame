package profile

type NodeCfg struct {
	NodeId   string         `json:"node_id"`
	NodeType string         `json:"node_type"`
	Address  map[string]int `json:"address"`
}

func (n *NodeCfg) ListenPorts() map[string]int {
	return n.Address
}

func (n *NodeCfg) GetNodeId() string {
	return n.NodeId
}

func (n *NodeCfg) GetNodeType() string {
	return n.NodeType
}
