package xconnector

import (
	"net"
	"sync"
	"testing"
)

func TestNewTCPConnector(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	tcp := NewTCP(":9071")
	tcp.OnConnect(func(conn net.Conn) {
		log.Infof("new pkg.Conn = %s", conn.RemoteAddr())
	})

	tcp.Start()

	wg.Wait()
}
