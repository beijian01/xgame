package cherryCluster

import (
	"bytes"
	"github.com/beijian01/xgame/framework/net/packet"
	"github.com/beijian01/xgame/framework/net/parser"
	"github.com/beijian01/xgame/pb"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"time"

	cerr "github.com/beijian01/xgame/framework/error"
	cfacade "github.com/beijian01/xgame/framework/facade"

	"github.com/nats-io/nats.go"
)

type (
	Cluster struct {
		app        cfacade.IApplication
		bufferSize int
		prefix     string
		natsSub    *natsSubject

		natsConn *NatsConn

		msgHandlers   *MessageHandlerMgr
		asyncCallback *rpcCallbackMgr
		responseWait  *ResponseWaitMgr
	}
)

func (p *Cluster) ListenMessage(cbk any) {
	p.msgHandlers.ListenMsg(cbk)
}

func (p *Cluster) PublishMsg(nodeId string, msg proto.Message) error {
	if !p.app.Running() {
		return cerr.ClusterRPCClientIsStop
	}

	nodeType, err := p.app.Discovery().GetType(nodeId)
	if err != nil {
		return err
	}
	subject := getRemoteSubject(p.prefix, nodeType, nodeId)

	common := &pb.MsgCommon{
		SourceId: p.app.GetNodeId(),
		TargetId: nodeId,
		MsgType:  pb.MsgType_SvrMsgTypPublish,
	}
	data, err := packet.PackMessage(common, msg)
	if err != nil {
		return err
	}

	return p.natsConn.Publish(subject, data)
}

func (p *Cluster) RequestWait(nodeId string, req proto.Message, timeout time.Duration) (proto.Message, error) {
	common := &pb.MsgCommon{
		SourceId: p.app.GetNodeId(),
		TargetId: nodeId,
		MsgType:  pb.MsgType_SvrMsgTypRequestWait,
		Mid:      p.responseWait.NextMid(),
	}
	data, err := packet.PackMessage(common, req)
	if err != nil {
		return nil, err
	}
	if err = p.SendBytes(nodeId, data); err != nil {
		return nil, err
	}
	return p.responseWait.WaitResponse(common.Mid, timeout)
}

func (p *Cluster) RequestAsync(nodeId string, req proto.Message, cbk func(resp proto.Message, err error)) error {
	//TODO implement me
	panic("implement me")
	return nil
}

func NewCluster(app cfacade.IApplication) cfacade.ICluster {
	cluster := &Cluster{
		app:        app,
		bufferSize: 1024,
		prefix:     "node",
		//natsSub:       newNatsSubject(getRemoteSubject()),
		//natsConn:      nil,
		msgHandlers:   NewMessageHandlerMgr(),
		asyncCallback: newRpcHandlerMgr(),
		responseWait:  newResponseWaitMgr(),
	}
	cluster.natsConn = NewNatsConn()
	cluster.natsSub = newNatsSubject(getRemoteSubject(cluster.prefix, app.GetNodeType(), app.GetNodeId()), cluster.bufferSize)

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

		// 将 natsMsg.RawMsg 解析成 parser.Message
		reader := bytes.NewReader(natsMsg.Data)
		common, msg, err := packet.ReadMessage(reader)
		if err != nil {
			logrus.Errorf("[receive] ReadMessage fail. [subject = %s, err = %v]", p.natsSub.subject, err)
			return
		}

		switch common.MsgType {
		case pb.MsgType_CliMsgTypRequest, pb.MsgType_CliMsgTypNotify:
			// client---->gate---->server
			// 当前服务节点是server
			// todo 在特定的worker中执行handler
			p.msgHandlers.reqHandlers[common.Route](common, msg)
		case pb.MsgType_CliMsgTypResponse, pb.MsgType_CliMsgTypPush:
			/// sever-> gate ->client
			// 当前节点是gate
			agents := p.app.Find(parser.AgentManagerComponentName).(*parser.AgentManager)
			agent, ok := agents.GetAgent(common.Sid)
			if ok {
				agent.Response(common) // gate 发给 client
			}
		case pb.MsgType_SvrMsgTypPublish:
			p.msgHandlers.reqHandlers[common.Route](common, msg)

		case pb.MsgType_SvrMsgTypRequestAsync:
			p.msgHandlers.reqHandlers[common.Route](common, msg)

		case pb.MsgType_SvrMsgTypResponseAsync:
			cbk := p.asyncCallback.getCallback(common.Mid)
			cbk(msg, nil)
		case pb.MsgType_SvrMsgTypRequestWait:
			handler, exist := p.msgHandlers.reqHandlers[common.Route]
			if !exist {
				logrus.Errorf("[receive] RequestWait fail. [subject = %s, err = %v]", p.natsSub.subject, err)
				return
			}
			handler(common, msg)
		case pb.MsgType_SvrMsgTypResponseWait:
			p.responseWait.pbChan[common.Mid] <- msg
		}
	}

	for msg := range p.natsSub.ch {
		process(msg)
	}
}

func (p *Cluster) SendBytes(nodeId string, data []byte) error {
	if !p.app.Running() {
		return cerr.ClusterRPCClientIsStop
	}
	nodeType, err := p.app.Discovery().GetType(nodeId)
	if err != nil {
		return err
	}
	subject := getRemoteSubject(p.prefix, nodeType, nodeId)
	return p.natsConn.Publish(subject, data)
}
