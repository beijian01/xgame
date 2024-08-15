package cherryCluster

import (
	"github.com/beijian01/xgame/framework/net/parser"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

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

		cliHandlers CliAgentHandlerMgr
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

		// todo 将 natsMsg.RawMsg 解析成 parser.CliMessage
		parserMsg := &parser.CliMessage{}
		var pbMsg proto.Message // todo 将parser.Message的pb data 反序列化成 proto.CliMessage
		switch parserMsg.MType {
		case parser.CliMsgTypRequest:
			// client->gate->server
			// todo 在特定的worker中执行handler
			p.cliHandlers.handlers[parserMsg.Route](parserMsg.Session, pbMsg)
		case parser.CliMsgTypResponse:
			/// sever-> gate ->client
			agents := p.app.Find(parser.AgentManagerComponentName).(*parser.AgentManager)
			agent, ok := agents.GetAgent(parserMsg.string)
			if ok {
				agent.Response(parserMsg)
			}
		case parser.CliMsgTypPush:

		case parser.CliMsgTypNotify:
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
