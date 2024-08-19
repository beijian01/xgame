package main

import (
	cherry "github.com/beijian01/xgame/framework"
	"github.com/beijian01/xgame/framework/facade"
	"github.com/beijian01/xgame/framework/net/xagent"
	"github.com/beijian01/xgame/framework/profile"
	"github.com/beijian01/xgame/pb"
	"github.com/sirupsen/logrus"
	"math/rand/v2"
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
			Ports:    map[string]int{"tcp": 1234},
		}, true)

		gate1.Find(xagent.AgentsName).(*xagent.Agents).RouteMessage((*pb.LoginRequest)(nil), "gate")
		cluster := gate1.Cluster()
		cluster.ListenMessage(func(sender facade.ISender, req *pb.LoginRequest) {
			logrus.Info("gate1 收到消息", req)
			sender.Resp(&pb.LoginResponse{
				Uid: rand.Uint64(),
			})
		})

		gate1.Startup()

	}()
	wg.Wait()
}
