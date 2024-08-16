package main

import (
	cherry "github.com/beijian01/xgame/framework"
	cherryCluster "github.com/beijian01/xgame/framework/net/cluster"
	cherryDiscovery "github.com/beijian01/xgame/framework/net/discovery"
	"github.com/beijian01/xgame/framework/net/packet"
	"github.com/beijian01/xgame/framework/net/profile"
	"github.com/beijian01/xgame/pb"
	"github.com/sirupsen/logrus"
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
			Address:  "",
		}, false)

		go func() {
			time.Sleep(time.Second * 3)
			respPb, err := gate1.Cluster().RequestWait("game1", &pb.GtGaReqAB{
				A: 1,
				B: 3,
			}, time.Second*10)
			if err != nil {
				panic(err)
			}
			logrus.Info(respPb.(*pb.GtGaRspAB))
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
			Address:  "",
		}, false)
		cluster := cherryCluster.New()
		game1.Register(cluster)
		game1.SetCluster(cluster)
		discovery := cherryDiscovery.New()
		game1.Register(discovery)
		game1.SetDiscovery(discovery)

		cluster.DoOnAfterInit = append(cluster.DoOnAfterInit, func() {
			game1.Cluster().ListenRequest(func(ext *pb.SvrExtend, req *pb.GtGaReqAB) {
				resp := &pb.GtGaRspAB{
					A:   req.A,
					B:   req.B,
					Sum: req.A + req.B,
				}
				data, err := packet.PackSvrMsg(&packet.SvrMessage{PBMsg: resp, PBExt: ext})
				if err != nil {
					panic(err)
				}
				game1.Cluster().SendBytes(ext.SourceId, data)
			})
		})

		game1.Startup()
	}()
	wg.Wait()
}
