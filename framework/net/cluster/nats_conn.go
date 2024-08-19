package xcluster

import (
	"github.com/beijian01/xgame/framework/profile"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/nats-io/nats.go"
)

type (
	NatsConn struct {
		*nats.Conn
		*profile.NatsCfg
		running bool
	}
)

func NewNatsConn(cfg *profile.NatsCfg) *NatsConn {
	conn := &NatsConn{
		NatsCfg: cfg,
	}
	if conn.Address == "" {
		conn.Address = nats.DefaultURL
	}
	return conn
}

func (p *NatsConn) Connect() {
	if p.running {
		return
	}

	for {
		conn, err := nats.Connect(p.Address)
		if err != nil {
			logrus.Warnf("nats connect fail! retrying in 3 seconds. err = %s", err)
			time.Sleep(3 * time.Second)
			continue
		}

		p.Conn = conn
		p.running = true
		logrus.Infof("nats is connected! [address = %s]", p.Address)
		break
	}
}

func (p *NatsConn) Close() {
	if p.running {
		p.running = false
		p.Conn.Close()
		logrus.Infof("nats connect execute Close()")
	}
}

func (p *NatsConn) Request(subj string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	return p.Conn.Request(subj, data, timeout)
}

func (p *NatsConn) ChanExecute(subject string, msgChan chan *nats.Msg, process func(msg *nats.Msg)) {
	_, chanErr := p.ChanSubscribe(subject, msgChan)
	if chanErr != nil {
		logrus.Error("subscribe fail. [subject = %s, err = %s]", subject, chanErr)
		return
	}

	for msg := range msgChan {
		process(msg)
	}
}
