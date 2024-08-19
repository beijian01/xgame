package xconnector

import (
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"testing"
)

func TestNewTCPConnector(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	tcp := NewTCP(":9071")
	tcp.OnConnect(func(conn net.Conn) {
		logrus.Infof("new net.Conn = %s", conn.RemoteAddr())
	})

	tcp.Start()

	wg.Wait()
}
