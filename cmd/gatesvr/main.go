package main

import (
	"flag"
	log "github.com/beijian01/xgame/framework/logger"
	"github.com/beijian01/xgame/framework/profile"
	"github.com/beijian01/xgame/gate"
	"sync"
)

var conf = flag.String("conf", "profile.json", "config file")
var nodeId = flag.String("nodeId", "gate1", "nodeId")

func main() {

	flag.Parse()
	cfg, err := profile.ParseProfile(*conf)
	if err != nil {
		panic(err)
	}
	nodeCfg, exist := cfg.FindNode(*nodeId)
	if !exist {
		panic("nodeId not exist")
	}
	log.Init(nodeCfg.NodeId, nodeCfg.NodeType, &nodeCfg.Log)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		gate.Run(cfg, *nodeId)
	}()

	wg.Wait()
}
