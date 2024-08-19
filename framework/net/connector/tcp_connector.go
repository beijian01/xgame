package xconnector

import (
	cfacade "github.com/beijian01/xgame/framework/facade"
	"github.com/sirupsen/logrus"
)

type (
	TCPConnector struct {
		cfacade.Component
		Connector
		Options
	}
)

func (*TCPConnector) Name() string {
	return "tcp_connector"
}

func (t *TCPConnector) OnAfterInit() {
	go t.Start()
}

func (t *TCPConnector) OnStop() {
	t.Stop()
}

func NewTCP(address string, opts ...Option) *TCPConnector {
	if address == "" {
		logrus.Warn("Create tcp connector fail. GetAddress is null.")
		return nil
	}

	tcp := &TCPConnector{
		Options: Options{
			address:  address,
			certFile: "",
			keyFile:  "",
			chanSize: 256,
		},
	}

	for _, opt := range opts {
		opt(&tcp.Options)
	}

	tcp.Connector = NewConnector(tcp.chanSize)

	return tcp
}

func (t *TCPConnector) Start() {
	listener, err := t.GetListener(t.certFile, t.keyFile, t.address)
	if err != nil {
		logrus.Fatalf("failed to listen: %s", err)
	}

	logrus.Infof("Tcp connector listening at GetAddress %s", t.address)
	if t.certFile != "" || t.keyFile != "" {
		logrus.Infof("certFile = %s, keyFile = %s", t.certFile, t.keyFile)
	}

	t.Connector.Start()

	for t.Running() {
		conn, err := listener.Accept()
		if err != nil {
			logrus.Errorf("Failed to accept TCP connection: %s", err.Error())
			continue
		}

		t.InChan(conn)
	}
}

func (t *TCPConnector) Stop() {
	t.Connector.Stop()
}
