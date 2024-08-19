package main

import (
	cherry "github.com/beijian01/xgame/framework"
	cherry "github.com/beijian01/xgame/framework/facade"
	cherryCluster "github.com/beijian01/xgame/framework/net/cluster"
	cherryDiscovery "github.com/beijian01/xgame/framework/net/discovery"
	"github.com/beijian01/xgame/framework/net/profile"
	"github.com/beijian01/xgame/pb"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		gate1 := cherry.NewAppNode(&profile.NodeCfg{
			NodeId:   "gate1",
			NodeType: "gate",
			Ports:    "",
		}, false)

		go func() {
			time.Sleep(time.Second)
			gate1.Cluster().RegisterResponse((*pb.GtGaRspAB)(nil))
			logrus.Info(gate1.Discovery().Map())
			time.Sleep(time.Second)
			err := gate1.Cluster().RequestAsync("game", &pb.GtGaReqAB{
				A: 1,
				B: 3,
			}, func(resp proto.Message, err error) {
				logrus.Info("async cbk  ", resp.(*pb.GtGaRspAB))
			})
			if err != nil {
				logrus.Errorf("gate1 request err: %v", err)
				return
			}

		}()
		cluster := cherryCluster.New()
		gate1.Register(cluster)
		gate1.SetCluster(cluster)
		discovery := cherryDiscovery.New()
		gate1.Register(discovery)
		gate1.SetDiscovery(discovery)

		gate1.Startup()

	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		game1 := cherry.NewAppNode(&profile.NodeCfg{
			NodeId:   "game1",
			NodeType: "game",
			Ports:    "",
		}, false)
		cluster := cherryCluster.New()
		game1.Register(cluster)
		game1.SetCluster(cluster)
		discovery := cherryDiscovery.New()
		game1.Register(discovery)
		game1.SetDiscovery(discovery)

		cluster.DoOnAfterInit = append(cluster.DoOnAfterInit, func() {
			game1.Cluster().ListenMessage(func(sender cherryFacade.ISender, req *pb.GtGaReqAB) {
				resp := &pb.GtGaRspAB{
					A:   req.A,
					B:   req.B,
					Sum: req.A + req.B,
				}
				sender.Resp(resp)
			})
		})
		go func() {
			time.Sleep(time.Second)
			logrus.Info(game1.Discovery().GetMember("gate1"))
		}()
		game1.Startup()
	}()
	wg.Wait()
}
