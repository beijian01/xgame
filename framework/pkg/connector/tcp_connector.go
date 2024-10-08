package xconnector

import (
	"github.com/beijian01/xgame/framework/facade"
	log "github.com/beijian01/xgame/framework/logger"
)

type (
	TCPConnector struct {
		facade.Component
		Connector
		Options
	}
)

func (*TCPConnector) Name() string {
	return "tcp_connector"
}

func (t *TCPConnector) OnAfterInit() {
}

func (t *TCPConnector) OnStop() {
	t.Stop()
}

func NewTCP(address string, opts ...Option) *TCPConnector {
	if address == "" {
		log.Warn("Create tcp connector fail. GetListenPorts is null.")
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
		log.Fatalf("failed to listen: %s", err)
	}

	log.Infof("Tcp connector listening at GetListenPorts %s", t.address)
	if t.certFile != "" || t.keyFile != "" {
		log.Infof("certFile = %s, keyFile = %s", t.certFile, t.keyFile)
	}

	t.Connector.Start()

	for t.Running() {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("Failed to accept TCP connection: %s", err.Error())
			continue
		}

		t.InChan(conn)
	}
}

func (t *TCPConnector) Stop() {
	t.Connector.Stop()
}
