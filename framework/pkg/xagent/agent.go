package xagent

import (
	"github.com/beijian01/xgame/framework/facade"
	log "github.com/beijian01/xgame/framework/logger"
	"github.com/beijian01/xgame/framework/pkg/packet"
	"github.com/beijian01/xgame/framework/util"
	"github.com/beijian01/xgame/pb"

	"google.golang.org/protobuf/proto"
	"net"
	"sync/atomic"
	"time"
)

const (
	AgentInit   int32 = 0
	AgentClosed int32 = 3
)

type (
	Agent struct {
		facade.IApplication             // app
		conn                net.Conn    // low-level conn fd
		state               int32       // current agent state
		session             *pb.Session // session
		chDie               chan struct {
		} // wait for close
		//chPending   chan *packet. // push message queue
		chWrite     chan []byte   // push bytes queue
		lastAt      int64         // last heartbeat unix time stamp
		onCloseFunc []OnCloseFunc // on close agent

		agentMgr  *Agents
		chPending chan *pendingMsg
	}

	OnCloseFunc func(*Agent)
)

func NewAgent(app facade.IApplication, conn net.Conn, session *pb.Session) *Agent {
	agent := &Agent{
		IApplication: app,
		conn:         conn,
		state:        AgentInit,
		session:      session,
		chDie:        make(chan struct{}),
		chPending:    make(chan *pendingMsg, writeBacklog),
		chWrite:      make(chan []byte, writeBacklog),
		lastAt:       0,
		onCloseFunc:  nil,
	}

	agent.session.Ip = agent.RemoteAddr()
	agent.SetLastAt()

	log.Debugf("[sid = %s,uid = %d] Agent create., ip = %s]",
		agent.SID(),
		agent.UID(),
		agent.RemoteAddr(),
	)

	return agent
}

func (a *Agent) State() int32 {
	return a.state
}

func (a *Agent) SetState(state int32) bool {
	oldValue := atomic.SwapInt32(&a.state, state)
	return oldValue != state
}

func (a *Agent) Session() *pb.Session {
	return a.session
}

func (a *Agent) UID() uint64 {
	return a.session.Uid
}

func (a *Agent) SID() string {
	return a.session.Sid
}

func (a *Agent) Bind(uid uint64) error {
	return a.agentMgr.BindUID(a.SID(), uid)
}

func (a *Agent) Unbind() {
	a.agentMgr.Unbind(a.SID())
}

func (a *Agent) SetLastAt() {
	atomic.StoreInt64(&a.lastAt, time.Now().Unix())
}

func (a *Agent) SendRaw(bytes []byte) {
	a.chWrite <- bytes
}

func (a *Agent) Close() {
	if a.SetState(AgentClosed) {
		select {
		case <-a.chDie:
		default:
			close(a.chDie)
		}
	}
}

func (a *Agent) Start() {
	go a.writeChan()
	go a.readChan()
}

func (a *Agent) readChan() {
	defer func() {
		log.Debugf("[sid = %s,uid = %d] Agent read chan exit.",
			a.SID(),
			a.UID(),
		)
		a.Close()
	}()

	for {
		common, msg, err := packet.ReadMessage(a.conn)
		if err != nil {
			log.Error(err)
			return
		}
		a.processPacket(common, msg)
	}
}

func (a *Agent) writeChan() {
	ticker := time.NewTicker(heartbeatTime)
	defer func() {
		log.Debugf("[sid = %s,uid = %d] Agent write chan exit.", a.SID(), a.UID())

		ticker.Stop()
		a.closeProcess()
		a.Close()
	}()

	for {
		select {
		case <-a.chDie:
			{
				return
			}
		case <-ticker.C:
			{
				deadline := time.Now().Add(-heartbeatTime).Unix()
				if a.lastAt < deadline {
					log.Debugf("[sid = %s,uid = %d] Check heartbeat timeout.", a.SID(), a.UID())
					return
				}
			}
		case pending := <-a.chPending:
			{
				a.processPending(pending.common, pending.msg)
			}
		case bytes := <-a.chWrite:
			{
				a.write(bytes)
			}
		}
	}
}

func (a *Agent) closeProcess() {
	util.Try(func() {
		for _, fn := range a.onCloseFunc {
			fn(a)
		}
	}, func(errString string) {
		log.Warn(errString)
	})

	a.Unbind()

	if err := a.conn.Close(); err != nil {
		log.Debugf("[sid = %s,uid = %d] Agent connect closed. [error = %s]",
			a.SID(),
			a.UID(),
			err,
		)
	}

	log.Debugf("[sid = %s,uid = %d] Agent closed. [count = %d, ip = %s]",
		a.SID(),
		a.UID(),
		a.agentMgr.Count(),
		a.RemoteAddr(),
	)

	//close(a.chPending)
	close(a.chWrite)
}

func (a *Agent) write(bytes []byte) {
	_, err := a.conn.Write(bytes)
	if err != nil {
		log.Warn(err)
	}
}

func (a *Agent) processPacket(common *pb.MsgCommon, msg proto.Message) {

	member, ok := a.Discovery().Random(a.agentMgr.pbRoute.route2nodeTyp[common.Route])
	if !ok {
		log.Warnf("[sid = %s,uid = %d] Node not found. [route = %d]",
			a.SID(),
			a.UID(),
			common.Route,
		)
		return
	}

	common.SourceId = a.GetNodeId()
	common.TargetId = member.GetNodeId()
	common.Uid = a.UID()
	common.Sid = a.SID()
	data, err := packet.PackMessage(common, msg)
	if err != nil {
		log.Errorf("pack svr msg error. [error = %s]", err)
		return
	}

	err = a.Cluster().SendBytes(member.GetNodeId(), data)
	if err != nil {
		log.Warnf("[sid = %s,uid = %d] Send bytes error. [error = %s]",
			a.SID(),
			a.UID(),
			err,
		)
		return
	}
	// update last time
	a.SetLastAt()
}

func (a *Agent) RemoteAddr() string {
	if a.conn != nil {
		return a.conn.RemoteAddr().String()
	}

	return ""
}

func (a *Agent) processPending(common *pb.MsgCommon, msg proto.Message) {
	// encode packet
	pkg, err := packet.PackMessage(common, msg)
	if err != nil {
		log.Warn(err)
		return
	}

	a.SendRaw(pkg)
}

func (a *Agent) sendPending(common *pb.MsgCommon, msg proto.Message) {
	if a.state == AgentClosed {
		log.Warnf("[sid = %s,uid = %d] Session is closed. [message=%#v]",
			a.SID(),
			a.UID(),
			msg,
		)
		return
	}

	if len(a.chPending) >= writeBacklog {
		log.Warnf("[sid = %s,uid = %d] send buffer exceed. [%#v]",
			a.SID(),
			a.UID(),
			msg,
		)
		return
	}

	a.chPending <- &pendingMsg{
		common: common,
		msg:    msg,
	}
}

func (a *Agent) AddOnClose(fn OnCloseFunc) {
	if fn != nil {
		a.onCloseFunc = append(a.onCloseFunc, fn)
	}
}

func (a *Agent) Response(common *pb.MsgCommon, msg proto.Message) {
	a.sendPending(common, msg)
}
