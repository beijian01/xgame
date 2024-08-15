package cherryCluster

import (
	"github.com/beijian01/xgame/framework/net/packet"
	"github.com/beijian01/xgame/framework/net/parser"
	"github.com/beijian01/xgame/pb"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	cerr "github.com/beijian01/xgame/framework/error"
	cfacade "github.com/beijian01/xgame/framework/facade"

	"github.com/nats-io/nats.go"
)

type ResponseWaitMgr struct {
	pbChan map[uint32]chan proto.Message
}

type (
	Cluster struct {
		app        cfacade.IApplication
		bufferSize int
		prefix     string
		natsSub    *natsSubject

		natsConn *NatsConn

		msgHandlers   MessageHandlerMgr
		asyncCallback rpcCallbackMgr
		responseWait  ResponseWaitMgr
	}
)

func NewCluster(app cfacade.IApplication) cfacade.ICluster {
	cluster := &Cluster{
		app:        app,
		bufferSize: 1024,
	}

	return cluster
}

func (p *Cluster) Init() {
	p.natsConn.Connect()

	go p.receive()

	logrus.Info("nats cluster execute OnInit().")
}

func (p *Cluster) Stop() {
	p.natsSub.stop()

	p.natsConn.Close()

	logrus.Info("nats cluster execute OnStop().")
}

// 持续接收集群消息并处理
func (p *Cluster) receive() {
	var err error
	p.natsSub.subscription, err = p.natsConn.ChanSubscribe(p.natsSub.subject, p.natsSub.ch)
	if err != nil {
		logrus.Errorf("[receive] Subscribe fail. [subject = %s, err = %s]", p.natsSub.subject, err)
		return
	}

	process := func(natsMsg *nats.Msg) {
		if dropped, err := p.natsSub.subscription.Dropped(); err != nil {
			logrus.Errorf("[receive] Dropped messages. [subject = %s, dropped = %d, err = %v]",
				p.natsSub.subject,
				dropped,
				err,
			)
		}

		// 将 natsMsg.RawMsg 解析成 parser.SvrMessage

		svrMsg, err := packet.ParseSvrMessage(natsMsg.Data)
		if err != nil {
			logrus.Errorf("[receive] ParseSvrMessage fail. [subject = %s, err = %v]", p.natsSub.subject, err)
			return
		}

		switch svrMsg.PBExt.MsgType {
		case pb.MsgType_CliMsgTypRequest, pb.MsgType_CliMsgTypNotify:
			// client---->gate---->server
			// 当前服务节点是server
			// todo 在特定的worker中执行handler
			// todo 封装session
			p.msgHandlers.cliHandlers[svrMsg.Route](&pb.Session{
				Sid:    svrMsg.PBExt.Sid,
				Uid:    svrMsg.PBExt.Uid,
				GateId: svrMsg.PBExt.SourceId,
			}, svrMsg.PBMsg)
		case pb.MsgType_CliMsgTypResponse, pb.MsgType_CliMsgTypPush:
			/// sever-> gate ->client
			// 当前节点是gate
			agents := p.app.Find(parser.AgentManagerComponentName).(*parser.AgentManager)
			agent, ok := agents.GetAgent(svrMsg.PBExt.Sid)
			if ok {
				agent.Response(svrMsg) // gate 发给 client
			}
		case pb.MsgType_SvrMsgTypPublish:

			// todo 第一个参数改server
			p.msgHandlers.svrHandlers[svrMsg.Route](&pb.Session{
				Sid:    svrMsg.PBExt.Sid,
				Uid:    svrMsg.PBExt.Uid,
				GateId: svrMsg.PBExt.SourceId,
			}, svrMsg.PBMsg)

		case pb.MsgType_SvrMsgTypRequestAsync:
			// todo 第一个参数改server
			p.msgHandlers.svrHandlers[svrMsg.Route](&pb.Session{
				Sid:    svrMsg.PBExt.Sid,
				Uid:    svrMsg.PBExt.Uid,
				GateId: svrMsg.PBExt.SourceId,
			}, svrMsg.PBMsg)

		case pb.MsgType_SvrMsgTypResponseAsync:
			cbk := p.asyncCallback.getCallback(svrMsg.PBExt.Mid)
			cbk(svrMsg.PBMsg, nil)
		case pb.MsgType_SvrMsgTypRequestWait:
			// todo 第一个参数改server
			p.msgHandlers.svrHandlers[svrMsg.Route](&pb.Session{
				Sid:    svrMsg.PBExt.Sid,
				Uid:    svrMsg.PBExt.Uid,
				GateId: svrMsg.PBExt.SourceId,
			}, svrMsg.PBMsg)

		case pb.MsgType_SvrMsgTypResponseWait:
			p.responseWait.pbChan[svrMsg.PBExt.Mid] <- svrMsg.PBMsg
		}
	}

	for msg := range p.natsSub.ch {
		process(msg)
	}
}

func (p *Cluster) SendMsg(nodeId string, request proto.Message) error {
	if !p.app.Running() {
		return cerr.ClusterRPCClientIsStop
	}

	nodeType, err := p.app.Discovery().GetType(nodeId)
	if err != nil {
		return err
	}
	subject := getRemoteSubject(p.prefix, nodeType, nodeId)
	// todo 按照约定格式组装消息
	bytes, err := proto.Marshal(request)
	if err != nil {
		logrus.Warn(err)
		return err
	}

	return p.natsConn.Publish(subject, bytes)
}

func (p *Cluster) PublishBytes(nodeId string, data []byte) error {
	if !p.app.Running() {
		return cerr.ClusterRPCClientIsStop
	}
	nodeType, err := p.app.Discovery().GetType(nodeId)
	if err != nil {
		return err
	}
	subject := getRemoteSubject(p.prefix, nodeType, "")
	return p.natsConn.Publish(subject, data)
}
