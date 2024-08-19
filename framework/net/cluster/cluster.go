package xcluster

import (
	"bytes"
	"github.com/beijian01/xgame/framework/net/packet"
	"github.com/beijian01/xgame/pb"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"time"

	cerr "github.com/beijian01/xgame/framework/error"
	"github.com/beijian01/xgame/framework/facade"

	"github.com/nats-io/nats.go"
)

type (
	Cluster struct {
		app        facade.IApplication
		bufferSize int
		prefix     string
		natsSub    *natsSubject

		natsConn *NatsConn

		msgHandlers *MessageHandlerMgr
	}
)

func (p *Cluster) ListenMessage(cbk any) {
	p.msgHandlers.ListenMsg(cbk)
}
func (p *Cluster) RegisterResponse(resp proto.Message) {
	p.msgHandlers.RegisterResponse(resp)
}
func (p *Cluster) SetDefaultHandler(handler facade.ReqMsgHandler) {
	p.msgHandlers.setDefaultHandler(handler)
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
	}
	data, err := packet.PackMessage(common, msg)
	if err != nil {
		return err
	}

	return p.natsConn.Publish(subject, data)
}

func (p *Cluster) RequestWait(nodeType string, req proto.Message, timeout time.Duration) (proto.Message, error) {
	return p.msgHandlers.requester.requestWait(nodeType, req, timeout)
}

func (p *Cluster) RequestAsync(nodeType string, req proto.Message, cbk func(resp proto.Message, err error)) error {
	return p.msgHandlers.requester.requestAsync(nodeType, req, cbk)
}

func newCluster(app facade.IApplication) facade.ICluster {
	cluster := &Cluster{
		app:         app,
		bufferSize:  1024,
		prefix:      "node",
		msgHandlers: NewMessageHandlerMgr(app),
	}
	cluster.natsConn = NewNatsConn(app.Profile().Nats)
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

		// 将 natsMsg.RawMsg 解析成 agent.Message
		reader := bytes.NewReader(natsMsg.Data)
		common, msg, err := packet.ReadMessage(reader)
		if err != nil {
			logrus.Errorf("[receive] ReadMessage fail. [subject = %s, err = %v]", p.natsSub.subject, err)
			return
		}
		sender := &Sender{
			MsgCommon: common,
			app:       p.app,
		}
		if handler, exist := p.msgHandlers.reqHandlers[common.Route]; exist {
			handler(sender, msg)
		} else {
			p.msgHandlers.defaultHandler(sender, msg)
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
