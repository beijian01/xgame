package cherryDiscovery

import (
	"context"
	"fmt"
	cfacade "github.com/beijian01/xgame/framework/facade"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/namespace"

	"strings"
	"time"
)

var (
	keyPrefix         = "/cherry/node/"
	registerKeyFormat = keyPrefix + "%s"
)

// ETCD etcd方式发现服务
type ETCD struct {
	app cfacade.IApplication
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

func (p *ETCD) Load(app cfacade.IApplication) {
	p.DiscoveryDefault.PreInit()
	p.app = app
	p.ttl = 10

	// todo etcd 配置加载
	p.init()
	p.getLeaseId()
	p.register()
	p.watch()

	logrus.Infof("[etcd] init complete! [endpoints = %v] [leaseId = %d]", p.config.Endpoints, p.leaseID)
}

func (p *ETCD) OnStop() {
	key := fmt.Sprintf(registerKeyFormat, p.app.NodeId())
	_, err := p.cli.Delete(context.Background(), key)
	logrus.Infof("etcd stopping! err = %v", err)

	err = p.cli.Close()
	if err != nil {
		logrus.Warnf("etcd stopping error! err = %v", err)
	}
}

func getDialTimeout(config jsoniter.Any) time.Duration {
	t := time.Duration(config.Get("dial_timeout_second").ToInt64()) * time.Second
	if t < 1*time.Second {
		t = 3 * time.Second
	}

	return t
}

func getEndPoints(config jsoniter.Any) []string {
	return strings.Split(config.Get("end_points").ToString(), ",")
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
	registerMember := &cproto.Member{
		NodeId:   p.app.NodeId(),
		NodeType: p.app.NodeType(),
		Address:  p.app.RpcAddress(),
		Settings: make(map[string]string),
	}

	jsonString, err := jsoniter.MarshalToString(registerMember)
	if err != nil {
		logrus.Fatal(err)
		return
	}

	key := fmt.Sprintf(registerKeyFormat, p.app.NodeId())
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
	member := &cproto.Member{}
	err := jsoniter.Unmarshal(data, member)
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
