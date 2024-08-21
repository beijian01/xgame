package main

import (
	"flag"
	"github.com/beijian01/xgame/auth"
	log "github.com/beijian01/xgame/framework/logger"
	"github.com/beijian01/xgame/framework/profile"
	"github.com/beijian01/xgame/gate"
	"sync"
)

var conf = flag.String("conf", "profile.json", "config file")

func main() {

	flag.Parse()
	cfg, err := profile.ParseProfile(*conf)
	if err != nil {
		panic(err)
	}
	log.Init(&cfg.Log)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		gate.Run(cfg, "gate1")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		auth.Run(cfg, "auth1")
	}()

	wg.Wait()
}
