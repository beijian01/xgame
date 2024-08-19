package xagent

import (
	"fmt"
	"github.com/beijian01/xgame/framework/facade"
	xconnector "github.com/beijian01/xgame/framework/net/connector"
	"github.com/beijian01/xgame/pb"
	"github.com/nats-io/nuid"
	"github.com/sirupsen/logrus"
	"net"
)

const NetParserName = "net_parser"

type (
	Parser struct {
		facade.Component

		connectors     []facade.IConnector
		onNewAgentFunc OnNewAgentFunc
		onInitFunc     func()

		agents *Agents
	}

	OnNewAgentFunc func(newAgent *Agent)
)

func NewNetParser(app facade.IApplication) *Parser {
	p := &Parser{}
	p.AddConnector(xconnector.NewTCP(fmt.Sprintf(":%d", app.GetListenPorts()["tcp"])))

	p.AddConnector(xconnector.NewWS(fmt.Sprintf(":%d", app.GetListenPorts()["ws"])))

	p.SetOnNewAgent(func(newAgent *Agent) {
		newAgent.AddOnClose(func(agent *Agent) {
			logrus.Infof("agent closed, sid: %s", agent.SID())
		})
	})

	return p
}

func (p *Parser) Name() string {
	return NetParserName
}
func (p *Parser) Init() {

	if len(p.connectors) < 1 {
		panic("connectors is nil. Please call the AddConnector(...) method add IConnector.")
	}

	p.agents = p.App().Find(AgentsName).(*Agents)

}

func (p *Parser) OnAfterInit() {
	for _, connector := range p.connectors {
		connector.OnConnect(p.defaultOnConnectFunc)
		go connector.Start() // start connector!
	}
}

func (p *Parser) AddConnector(connector facade.IConnector) {
	p.connectors = append(p.connectors, connector)
}

func (p *Parser) Connectors() []facade.IConnector {
	return p.connectors
}

func (p *Parser) defaultOnConnectFunc(conn net.Conn) {
	session := &pb.Session{
		Sid: nuid.Next(),
	}

	agent := NewAgent(p.App(), conn, session)

	if p.onNewAgentFunc != nil {
		p.onNewAgentFunc(agent)
	}

	p.agents.BindSID(agent)
	agent.Run()
}

func (p *Parser) SetOnNewAgent(fn func(newAgent *Agent)) {
	p.onNewAgentFunc = fn
}
