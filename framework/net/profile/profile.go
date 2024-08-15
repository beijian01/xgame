package profile

type NodeCfg struct {
	NodeId   string `json:"node_id"`
	NodeType string `json:"node_type"`
	Address  string `json:"address"`
}

func (n *NodeCfg) GetNodeId() string {
	return n.NodeId
}

func (n *NodeCfg) GetNodeType() string {
	return n.NodeType
}

func (n *NodeCfg) GetAddress() string {
	return n.Address
}
