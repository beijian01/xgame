package cherryConnector

import (
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"

	cfacade "github.com/beijian01/xgame/framework/facade"
	"github.com/gorilla/websocket"
)

type (
	WSConnector struct {
		cfacade.Component
		Connector
		Options
		upgrade *websocket.Upgrader
	}

	// WSConn is an adapter to t.INetConn, which implements all INetConn
	// interface base on *websocket.INetConn
	WSConn struct {
		*websocket.Conn
		typ    int // message type
		reader io.Reader
	}
)

func (*WSConnector) Name() string {
	return "websocket_connector"
}

func (w *WSConnector) OnAfterInit() {
}

func (w *WSConnector) OnStop() {
	w.Stop()
}

func NewWS(address string, opts ...Option) *WSConnector {
	if address == "" {
		logrus.Warn("create websocket fail. address is null.")
		return nil
	}

	ws := &WSConnector{
		Options: Options{
			address:  address,
			certFile: "",
			keyFile:  "",
			chanSize: 256,
		},
		upgrade: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		},
	}

	for _, opt := range opts {
		opt(&ws.Options)
	}

	ws.Connector = NewConnector(ws.chanSize)

	return ws
}

func (w *WSConnector) Start() {
	listener, err := w.GetListener(w.certFile, w.keyFile, w.address)
	if err != nil {
		logrus.Fatalf("failed to listen: %s", err)
	}

	logrus.Infof("Websocket connector listening at GetAddress %s", w.address)
	if w.certFile != "" || w.keyFile != "" {
		logrus.Infof("certFile = %s, keyFile = %s", w.certFile, w.keyFile)
	}

	w.Connector.Start()

	http.Serve(listener, w)
}

func (w *WSConnector) Stop() {
	w.Connector.Stop()
}

func (w *WSConnector) SetUpgrade(upgrade *websocket.Upgrader) {
	if upgrade != nil {
		w.upgrade = upgrade
	}
}

func (w *WSConnector) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	wsConn, err := w.upgrade.Upgrade(rw, r, nil)
	if err != nil {
		logrus.Infof("Upgrade failure, URI=%s, Error=%s", r.RequestURI, err.Error())
		return
	}

	conn := NewWSConn(wsConn)
	w.InChan(&conn)
}

// NewWSConn return an initialized *WSConn
func NewWSConn(conn *websocket.Conn) WSConn {
	c := WSConn{
		Conn: conn,
	}
	return c
}

func (c *WSConn) Read(b []byte) (int, error) {
	if c.reader == nil {
		t, r, err := c.NextReader()
		if err != nil {
			return 0, err
		}
		c.typ = t
		c.reader = r
	}
	n, err := c.reader.Read(b)
	if err != nil && err != io.EOF {
		return n, err
	} else if err == io.EOF {
		_, r, err := c.NextReader()
		if err != nil {
			return 0, err
		}
		c.reader = r
	}

	return n, nil
}

func (c *WSConn) Write(b []byte) (int, error) {
	err := c.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

func (c *WSConn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}

	return c.SetWriteDeadline(t)
}
