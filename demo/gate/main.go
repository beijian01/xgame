package main

import (
	"fmt"
	cherry "github.com/beijian01/xgame/framework"
	cfacade "github.com/beijian01/xgame/framework/facade"
	cherryCluster "github.com/beijian01/xgame/framework/net/cluster"
	cherryConnector "github.com/beijian01/xgame/framework/net/connector"
	cherryDiscovery "github.com/beijian01/xgame/framework/net/discovery"
	"github.com/beijian01/xgame/framework/net/profile"
	"github.com/beijian01/xgame/pb"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"sync"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		gate1 := cherry.NewAppNode(&profile.NodeCfg{
			NodeId:   "gate1",
			NodeType: "gate",
			Address:  "",
		}, true)

		cluster := cherryCluster.New()
		gate1.Register(cluster)
		gate1.SetCluster(cluster)
		discovery := cherryDiscovery.New()
		gate1.Register(discovery)
		gate1.SetDiscovery(discovery)

		agentsManager := xagent.NewAgents()
		gate1.Register(agentsManager)

		tcpConn := cherryConnector.NewTCP(":1234")
		tcpConn.OnConnect(func(conn net.Conn) {
			xagent1 := xagent.NewAgent(gate1, conn, &pb.Session{
				Uid: rand.Uint64(),
				Sid: fmt.Sprint("sid====", rand.Uint64()),
			})
			agentsManager.BindSID(xagent1)
			xagent1.Run()
		})

		gate1.Register(tcpConn)

		cluster.DoOnAfterInit = append(cluster.DoOnAfterInit, func() {
			agentsManager.RouteMessage(&pb.LoginRequest{}, "gate")
			gate1.Cluster().ListenMessage(func(sender cfacade.ISender, req *pb.LoginRequest) {
				logrus.Info("gate1 收到消息", req)
				sender.Resp(&pb.LoginResponse{
					Uid: rand.Uint64(),
				})
			})
		})
		//go func() {
		//	time.Sleep(3 * time.Second)
		//	logrus.Info(gate1.Discovery().Map())
		//	logrus.Info(gate1.Discovery().ListByType("gate"))
		//}()
		gate1.Startup()

	}()
	wg.Wait()
}
