package xdiscovery

import (
	"github.com/beijian01/xgame/framework/facade"
	"math/rand"
)

func (n *DiscoveryDefault) Random(nodeType string) (facade.IMember, bool) {
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
