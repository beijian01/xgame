package main

import (
	cherry "github.com/beijian01/xgame/framework"
	"github.com/beijian01/xgame/framework/net/profile"
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
		game1.Startup()
	}()
	wg.Wait()
}
