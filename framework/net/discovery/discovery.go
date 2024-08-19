package xdiscovery

import (
	"github.com/beijian01/xgame/framework/util"
	"github.com/sirupsen/logrus"
	"math/rand"
	"sync"

	cerr "github.com/beijian01/xgame/framework/error"

	cfacade "github.com/beijian01/xgame/framework/facade"
)

type DiscoveryDefault struct {
	memberMap        sync.Map // key:nodeId,value:cfacade.IMember
	onAddListener    []cfacade.MemberListener
	onRemoveListener []cfacade.MemberListener
}

func (n *DiscoveryDefault) PreInit() {
	n.memberMap = sync.Map{}
}

func (n *DiscoveryDefault) Load(_ cfacade.IApplication) {
}

func (n *DiscoveryDefault) Name() string {
	return "default"
}

func (n *DiscoveryDefault) Map() map[string]cfacade.IMember {
	memberMap := map[string]cfacade.IMember{}

	n.memberMap.Range(func(key, value any) bool {
		if member, ok := value.(cfacade.IMember); ok {
			memberMap[member.GetNodeId()] = member
		}
		return true
	})

	return memberMap
}

func (n *DiscoveryDefault) ListByType(nodeType string, filterNodeId ...string) []cfacade.IMember {
	var memberList []cfacade.IMember

	n.memberMap.Range(func(key, value any) bool {
		member := value.(cfacade.IMember)
		if member.GetNodeType() == nodeType {
			if _, ok := util.StringIn(member.GetNodeId(), filterNodeId); !ok {
				memberList = append(memberList, member)
			}
		}

		return true
	})

	return memberList
}

func (n *DiscoveryDefault) Random(nodeType string) (cfacade.IMember, bool) {
	memberList := n.ListByType(nodeType)
	memberLen := len(memberList)

	if memberLen < 1 {
		return nil, false
	}

	if memberLen == 1 {
		return memberList[0], true
	}

	return memberList[rand.Intn(len(memberList))], true
}

func (n *DiscoveryDefault) GetType(nodeId string) (nodeType string, err error) {
	member, found := n.GetMember(nodeId)
	if !found {
		return "", cerr.Errorf("nodeId = %s not found.", nodeId)
	}
	return member.GetNodeType(), nil
}

func (n *DiscoveryDefault) GetMember(nodeId string) (cfacade.IMember, bool) {
	if nodeId == "" {
		return nil, false
	}

	value, found := n.memberMap.Load(nodeId)
	if !found {
		return nil, false
	}

	return value.(cfacade.IMember), found
}

func (n *DiscoveryDefault) AddMember(member cfacade.IMember) {
	_, loaded := n.memberMap.LoadOrStore(member.GetNodeId(), member)
	if loaded {
		logrus.Warnf("duplicate nodeId. [nodeType = %s], [nodeId = %s]",
			member.GetNodeType(),
			member.GetNodeId(),
		)
		return
	}

	for _, listener := range n.onAddListener {
		listener(member)
	}

	logrus.Debugf("addMember new member. [member = %s]", member)
}

func (n *DiscoveryDefault) RemoveMember(nodeId string) {
	value, loaded := n.memberMap.LoadAndDelete(nodeId)
	if loaded {
		member := value.(cfacade.IMember)
		logrus.Debugf("remove member. [member = %s]", member)

		for _, listener := range n.onRemoveListener {
			listener(member)
		}
	}
}

func (n *DiscoveryDefault) OnAddMember(listener cfacade.MemberListener) {
	if listener == nil {
		return
	}
	n.onAddListener = append(n.onAddListener, listener)
}

func (n *DiscoveryDefault) OnRemoveMember(listener cfacade.MemberListener) {
	if listener == nil {
		return
	}
	n.onRemoveListener = append(n.onRemoveListener, listener)
}

func (n *DiscoveryDefault) Stop() {

}
