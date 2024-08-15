package parser

import (
	cerr "github.com/beijian01/xgame/framework/error"
	cherryFacade "github.com/beijian01/xgame/framework/facade"
	"github.com/sirupsen/logrus"
	"sync"
)

const AgentManagerComponentName = "AgentManager"

type AgentManager struct {
	cherryFacade.IComponent

	lock        sync.RWMutex
	sidAgentMap map[string]*Agent // sid -> Agent
	uidMap      map[uint64]string // uid -> sid

	pbRoute *routeMgr
}

func NewAgents() *AgentManager {
	return &AgentManager{
		sidAgentMap: make(map[string]*Agent),
		uidMap:      make(map[uint64]string),
		pbRoute:     newRouteMgr(),
	}
}

func (a *AgentManager) BindSID(agent *Agent) {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.sidAgentMap[agent.SID()] = agent
}

func (a *AgentManager) BindUID(sid string, uid uint64) error {
	if sid == "" {
		return cerr.Errorf("[sid = %s] less than 1.", sid)
	}

	if uid < 1 {
		return cerr.Errorf("[uid = %d] less than 1.", uid)
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	agent, found := a.sidAgentMap[sid]
	if !found {
		return cerr.Errorf("[sid = %s] does not exist.", sid)
	}

	if agent.UID() > 0 && agent.UID() == uid {
		return cerr.Errorf("[uid = %d] has already bound.", agent.UID())
	}

	agent.session.Uid = uid
	a.uidMap[uid] = sid

	return nil
}

func (a *AgentManager) Unbind(sid string) {
	a.lock.Lock()
	defer a.lock.Unlock()

	agent, found := a.sidAgentMap[sid]
	if !found {
		return
	}

	delete(a.sidAgentMap, sid)
	delete(a.uidMap, agent.UID())

	sidCount := len(a.sidAgentMap)
	uidCount := len(a.uidMap)
	if sidCount == 0 || uidCount == 0 {
		logrus.Infof("Unbind agent sid = %s, sidCount = %d, uidCount = %d", sid, sidCount, uidCount)
	}
}

func (a *AgentManager) GetAgent(sid string) (*Agent, bool) {
	a.lock.Lock()
	defer a.lock.Unlock()

	agent, found := a.sidAgentMap[sid]
	return agent, found
}

func (a *AgentManager) GetAgentWithUID(uid uint64) (*Agent, bool) {
	if uid < 1 {
		return nil, false
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	sid, found := a.uidMap[uid]
	if !found {
		return nil, false
	}

	agent, found := a.sidAgentMap[sid]
	return agent, found
}

func (a *AgentManager) ForeachAgent(fn func(a *Agent)) {
	for _, agent := range a.sidAgentMap {
		fn(agent)
	}
}

func (a *AgentManager) Count() int {
	a.lock.RLock()
	defer a.lock.RUnlock()

	return len(a.sidAgentMap)
}
