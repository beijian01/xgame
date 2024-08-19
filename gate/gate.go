package gate

import (
	xgame "github.com/beijian01/xgame/framework"
	"github.com/beijian01/xgame/framework/facade"
	"github.com/beijian01/xgame/framework/net/xagent"
	"github.com/beijian01/xgame/framework/profile"
	"github.com/beijian01/xgame/pb"
)

//var conf = flag.String("conf", "profile.json", "config file")

func Run(cfg *profile.ClusterCfg, nodeId string) error {
	node, err := xgame.NewAppNode(cfg, nodeId)
	if err != nil {
		return err
	}
	routeMessage(node)
	listenMessage(node)
	node.Startup()
	return nil
}

func routeMessage(gate facade.IApplication) {
	agents := gate.Find(xagent.AgentsName).(*xagent.Agents)
	agents.RouteMessage((*pb.CReqLogin)(nil), profile.NodeTypeGate)
}

func listenMessage(gate facade.IApplication) {
	cluster := gate.Cluster()
	cluster.ListenMessage(onCReqLogin)
}
