package xdiscovery

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/beijian01/xgame/framework/facade"
	"github.com/beijian01/xgame/pb"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/namespace"

	"strings"
)

var (
	keyPrefix         = "/xgame/node/"
	registerKeyFormat = keyPrefix + "%s"
)

// ETCD etcd方式发现服务
type ETCD struct {
	app facade.IApplication
	DiscoveryDefault
	prefix  string
	config  clientv3.Config
	ttl     int64
	cli     *clientv3.Client // etcd client
	leaseID clientv3.LeaseID // get lease id
}

func NewDiscoveryETCD() *ETCD {
	return &ETCD{}
}

func (p *ETCD) Name() string {
	return "etcd"
}

func (p *ETCD) Load(app facade.IApplication) {
	p.DiscoveryDefault.PreInit()
	p.app = app
	p.ttl = 10
	p.prefix = app.Profile().Project // 使用项目名称作为etcd的namespace
	keyPrefix = fmt.Sprintf("/%s/node/", p.prefix)
	registerKeyFormat = keyPrefix + "%s"

	p.config.Endpoints = app.Profile().Etcd.Endpoints

	p.init()
	p.getLeaseId()
	p.register()
	p.watch()

	logrus.Infof("[etcd] init complete! [endpoints = %v] [leaseId = %d]", p.config.Endpoints, p.leaseID)
}

func (p *ETCD) OnStop() {
	key := fmt.Sprintf(registerKeyFormat, p.app.GetNodeId())
	_, err := p.cli.Delete(context.Background(), key)
	logrus.Infof("etcd stopping! err = %v", err)

	err = p.cli.Close()
	if err != nil {
		logrus.Warnf("etcd stopping error! err = %v", err)
	}
}

func (p *ETCD) init() {
	var err error
	p.cli, err = clientv3.New(p.config)
	if err != nil {
		logrus.Fatalf("etcd connect fail. err = %v", err)
		return
	}

	// set namespace
	p.cli.KV = namespace.NewKV(p.cli.KV, p.prefix)
	p.cli.Watcher = namespace.NewWatcher(p.cli.Watcher, p.prefix)
	p.cli.Lease = namespace.NewLease(p.cli.Lease, p.prefix)
}

func (p *ETCD) getLeaseId() {
	var err error
	//设置租约时间
	resp, err := p.cli.Grant(context.Background(), p.ttl)
	if err != nil {
		logrus.Fatal(err)
		return
	}

	p.leaseID = resp.ID

	//设置续租 定期发送需求请求
	keepaliveChan, err := p.cli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		logrus.Fatal(err)
		return
	}

	go func() {
		for {
			select {
			case <-keepaliveChan:
				{
				}
			case die := <-p.app.DieChan():
				{
					if die {
						return
					}
				}
			}
		}
	}()
}

func (p *ETCD) register() {
	registerMember := &pb.Member{
		NodeId:   p.app.GetNodeId(),
		NodeType: p.app.GetNodeType(),
	}

	jsonBytes, err := json.Marshal(registerMember)
	jsonString := string(jsonBytes)
	if err != nil {
		logrus.Fatal(err)
		return
	}

	key := fmt.Sprintf(registerKeyFormat, p.app.GetNodeId())
	_, err = p.cli.Put(context.Background(), key, jsonString, clientv3.WithLease(p.leaseID))
	if err != nil {
		logrus.Fatal(err)
		return
	}
}

func (p *ETCD) watch() {
	resp, err := p.cli.Get(context.Background(), keyPrefix, clientv3.WithPrefix())
	if err != nil {
		logrus.Fatal(err)
		return
	}

	for _, ev := range resp.Kvs {
		p.addMember(ev.Value)
	}

	watchChan := p.cli.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())
	go func() {
		for rsp := range watchChan {
			for _, ev := range rsp.Events {
				switch ev.Type {
				case mvccpb.PUT:
					{
						p.addMember(ev.Kv.Value)
					}
				case mvccpb.DELETE:
					{
						p.removeMember(ev.Kv)
					}
				}
			}
		}
	}()
}

func (p *ETCD) addMember(data []byte) {
	member := &pb.Member{}
	err := json.Unmarshal(data, member)
	if err != nil {
		return
	}

	p.AddMember(member)
}

func (p *ETCD) removeMember(kv *mvccpb.KeyValue) {
	key := string(kv.Key)
	nodeId := strings.ReplaceAll(key, keyPrefix, "")
	if nodeId == "" {
		logrus.Warn("remove member nodeId is empty!")
	}

	p.RemoveMember(nodeId)
}
