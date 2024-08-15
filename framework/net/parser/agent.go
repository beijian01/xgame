package parser

import (
	cfacade "github.com/beijian01/xgame/framework/facade"
	"github.com/beijian01/xgame/framework/net/packet"
	"github.com/beijian01/xgame/framework/util"
	"github.com/beijian01/xgame/pb"
	"github.com/sirupsen/logrus"
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
		cfacade.IApplication             // app
		conn                 net.Conn    // low-level conn fd
		state                int32       // current agent state
		session              *pb.Session // session
		chDie                chan struct {
		} // wait for close
		chPending   chan *packet.SvrMessage // push message queue
		chWrite     chan []byte             // push bytes queue
		lastAt      int64                   // last heartbeat unix time stamp
		onCloseFunc []OnCloseFunc           // on close agent

		agentMgr *AgentManager
	}

	OnCloseFunc func(*Agent)
)

func NewAgent(app cfacade.IApplication, conn net.Conn, session *pb.Session) Agent {
	agent := Agent{
		IApplication: app,
		conn:         conn,
		state:        AgentInit,
		session:      session,
		chDie:        make(chan struct{}),
		chPending:    make(chan *packet.SvrMessage, writeBacklog),
		chWrite:      make(chan []byte, writeBacklog),
		lastAt:       0,
		onCloseFunc:  nil,
	}

	agent.session.Ip = agent.RemoteAddr()
	agent.SetLastAt()

	logrus.Debugf("[sid = %s,uid = %d] Agent create. [count = %d, ip = %s]",
		agent.SID(),
		agent.UID(),
		agent.agentMgr.Count(),
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

func (a *Agent) Run() {
	go a.writeChan()
	go a.readChan()
}

func (a *Agent) readChan() {
	defer func() {
		logrus.Debugf("[sid = %s,uid = %d] Agent read chan exit.",
			a.SID(),
			a.UID(),
		)
		a.Close()
	}()

	for {
		msg, isBreak, err := packet.ReadCliMessage(a.conn)
		if isBreak || err != nil {
			return
		}

		a.processPacket(msg)
	}
}

func (a *Agent) writeChan() {
	ticker := time.NewTicker(heartbeatTime)
	defer func() {
		logrus.Debugf("[sid = %s,uid = %d] Agent write chan exit.", a.SID(), a.UID())

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
					logrus.Debugf("[sid = %s,uid = %d] Check heartbeat timeout.", a.SID(), a.UID())
					return
				}
			}
		case pending := <-a.chPending:
			{
				a.processPending(pending)
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
		logrus.Warn(errString)
	})

	a.Unbind()

	if err := a.conn.Close(); err != nil {
		logrus.Debugf("[sid = %s,uid = %d] Agent connect closed. [error = %s]",
			a.SID(),
			a.UID(),
			err,
		)
	}

	logrus.Debugf("[sid = %s,uid = %d] Agent closed. [count = %d, ip = %s]",
		a.SID(),
		a.UID(),
		a.agentMgr.Count(),
		a.RemoteAddr(),
	)

	close(a.chPending)
	close(a.chWrite)
}

func (a *Agent) write(bytes []byte) {
	_, err := a.conn.Write(bytes)
	if err != nil {
		logrus.Warn(err)
	}
}

func (a *Agent) processPacket(msg *packet.CliMessage) {
	// 转发客户端消息至目标服务

	// todo 减少序列化和反序列化
	svrMsg := &packet.SvrMessage{
		PBMsg: msg.PBMsg,
		PBExt: &pb.SvrExtend{
			SourceId: a.NodeId(),
			TargetId: a.agentMgr.pbRoute.route2nodeTyp[msg.Route],
			Mid:      msg.MID,
			Sid:      a.SID(),
			Uid:      a.UID(),
			MsgType:  pb.MsgType_CliMsgTypRequest,
		},
	}
	data, err := packet.PackSvrMsg(svrMsg)
	if err != nil {
		logrus.Errorf("pack svr msg error. [error = %s]", err)
		return
	}
	if member, ok := a.Discovery().Random(a.agentMgr.pbRoute.route2nodeTyp[msg.Route]); ok {
		a.Cluster().PublishBytes(member.GetNodeId(), data)
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

func (a *Agent) processPending(pending *packet.SvrMessage) {

	// encode packet
	pkg, err := packet.PackSvrMsg(pending)
	if err != nil {
		logrus.Warn(err)
		return
	}

	a.SendRaw(pkg)
}

func (a *Agent) sendPending(message *packet.SvrMessage) {
	if a.state == AgentClosed {
		logrus.Warnf("[sid = %s,uid = %d] Session is closed. [message=%#v]",
			a.SID(),
			a.UID(),
			message,
		)
		return
	}

	if len(a.chPending) >= writeBacklog {
		logrus.Warnf("[sid = %s,uid = %d] send buffer exceed. [%#v]",
			a.SID(),
			a.UID(),
			message,
		)
		return
	}

	a.chPending <- message
}

func (a *Agent) AddOnClose(fn OnCloseFunc) {
	if fn != nil {
		a.onCloseFunc = append(a.onCloseFunc, fn)
	}
}

func (a *Agent) Response(msg *packet.SvrMessage) {
	a.sendPending(msg)
}
