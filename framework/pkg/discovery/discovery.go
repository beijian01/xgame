package xdiscovery

import (
	log "github.com/beijian01/xgame/framework/logger"
	"github.com/beijian01/xgame/framework/util"

	"sync"

	cerr "github.com/beijian01/xgame/framework/error"

	"github.com/beijian01/xgame/framework/facade"
)

type DiscoveryDefault struct {
	memberMap        sync.Map // key:nodeId,value:facade.IMember
	onAddListener    []facade.MemberListener
	onRemoveListener []facade.MemberListener
}

func (n *DiscoveryDefault) PreInit() {
	n.memberMap = sync.Map{}
}

func (n *DiscoveryDefault) Load(_ facade.IApplication) {
}

func (n *DiscoveryDefault) Name() string {
	return "default"
}

func (n *DiscoveryDefault) Map() map[string]facade.IMember {
	memberMap := map[string]facade.IMember{}

	n.memberMap.Range(func(key, value any) bool {
		if member, ok := value.(facade.IMember); ok {
			memberMap[member.GetNodeId()] = member
		}
		return true
	})

	return memberMap
}

func (n *DiscoveryDefault) ListByType(nodeType string, filterNodeId ...string) []facade.IMember {
	var memberList []facade.IMember

	n.memberMap.Range(func(key, value any) bool {
		member := value.(facade.IMember)
		if member.GetNodeType() == nodeType {
			if _, ok := util.StringIn(member.GetNodeId(), filterNodeId); !ok {
				memberList = append(memberList, member)
			}
		}

		return true
	})

	return memberList
}

func (n *DiscoveryDefault) GetType(nodeId string) (nodeType string, err error) {
	member, found := n.GetMember(nodeId)
	if !found {
		return "", cerr.Errorf("nodeId = %s not found.", nodeId)
	}
	return member.GetNodeType(), nil
}

func (n *DiscoveryDefault) GetMember(nodeId string) (facade.IMember, bool) {
	if nodeId == "" {
		return nil, false
	}

	value, found := n.memberMap.Load(nodeId)
	if !found {
		return nil, false
	}

	return value.(facade.IMember), found
}

func (n *DiscoveryDefault) AddMember(member facade.IMember) {
	_, loaded := n.memberMap.LoadOrStore(member.GetNodeId(), member)
	if loaded {
		log.Warnf("duplicate nodeId. [nodeType = %s], [nodeId = %s]",
			member.GetNodeType(),
			member.GetNodeId(),
		)
		return
	}

	for _, listener := range n.onAddListener {
		listener(member)
	}

	log.Debugf("addMember new member. [member = %s]", member)
}

func (n *DiscoveryDefault) RemoveMember(nodeId string) {
	value, loaded := n.memberMap.LoadAndDelete(nodeId)
	if loaded {
		member := value.(facade.IMember)
		log.Debugf("remove member. [member = %s]", member)

		for _, listener := range n.onRemoveListener {
			listener(member)
		}
	}
}

func (n *DiscoveryDefault) OnAddMember(listener facade.MemberListener) {
	if listener == nil {
		return
	}
	n.onAddListener = append(n.onAddListener, listener)
}

func (n *DiscoveryDefault) OnRemoveMember(listener facade.MemberListener) {
	if listener == nil {
		return
	}
	n.onRemoveListener = append(n.onRemoveListener, listener)
}

func (n *DiscoveryDefault) Stop() {

}
