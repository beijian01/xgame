package cherryCluster

type Member struct {
	NodeId   string            `json:"nodeId,omitempty"`
	NodeType string            `json:"nodeType,omitempty"`
	Address  string            `json:"address,omitempty"`
	Settings map[string]string `json:"settings,omitempty"`
}

type MemberList struct {
	List []*Member `json:"list,omitempty"`
}
