package xagent

import (
	cerr "github.com/beijian01/xgame/framework/error"
	cherryFacade "github.com/beijian01/xgame/framework/facade"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"sync"
)

const ComponentName = "agents_manager"

type Manager struct {
	cherryFacade.Component

	lock        sync.RWMutex
	sidAgentMap map[string]*Agent // sid -> Agent
	uidMap      map[uint64]string // uid -> sid

	pbRoute *routeMgr
}

func NewAgents() *Manager {
	return &Manager{
		sidAgentMap: make(map[string]*Agent),
		uidMap:      make(map[uint64]string),
		pbRoute:     newRouteMgr(),
	}
}

func (a *Manager) Name() string {
	return ComponentName
}

func (a *Manager) BindSID(agent *Agent) {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.sidAgentMap[agent.SID()] = agent
	agent.agentMgr = a
}

func (a *Manager) BindUID(sid string, uid uint64) error {
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

func (a *Manager) Unbind(sid string) {
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

func (a *Manager) GetAgent(sid string) (*Agent, bool) {
	a.lock.Lock()
	defer a.lock.Unlock()

	agent, found := a.sidAgentMap[sid]
	return agent, found
}

func (a *Manager) GetAgentWithUID(uid uint64) (*Agent, bool) {
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

func (a *Manager) ForeachAgent(fn func(a *Agent)) {
	for _, agent := range a.sidAgentMap {
		fn(agent)
	}
}

func (a *Manager) Count() int {
	a.lock.RLock()
	defer a.lock.RUnlock()

	return len(a.sidAgentMap)
}

func (a *Manager) Init() {
	a.App().Cluster().SetDefaultHandler(func(ext cherryFacade.ISender, msg proto.Message) {
		agent, exist := a.GetAgent(ext.GetCommon().GetSid())
		if !exist {
			logrus.Errorf("[sid = %s] not exist.", ext.GetCommon().GetSid())
			return
		}
		agent.Response(ext.GetCommon(), msg)
	})
}

func (a *Manager) RouteMessage(msg proto.Message, nodeType string) {
	a.pbRoute.addRoute(msg, nodeType)
}
