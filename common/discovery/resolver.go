package discovery

import (
	"common/config"
	"common/logs"
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"time"
)

type Resolver struct {
	conf        config.EtcdConf
	etcdClient  *clientv3.Client // etcd 连接
	dialTimeout int              // 超时时间
	closeCh     chan struct{}    // 关闭标识
	key         string
	cc          resolver.ClientConn
	srcAddrList []resolver.Address
	watchCh     clientv3.WatchChan
}

// Build 当grpc.Dial启动时, 调用此方法
func (r Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	//todo 获取到调用的key (user/v1) 连接etcd 获取对应的 value
	// 1. 建立 etcd 的连接
	r.cc = cc
	var err error
	r.etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   r.conf.Addrs,
		DialTimeout: time.Duration(r.dialTimeout) * time.Second,
	})
	if err != nil {
		logs.Fatal("grpc client connect etcd err:%v", err)
	}
	r.closeCh = make(chan struct{})
	// 2. 根据 key 获取 value
	r.key = target.URL.Path
	if err := r.sync(); err != nil {
		return nil, err
	}
	// 3. 节点变动, 更新数据
	go r.watch()
	return nil, nil
}

func (r Resolver) Scheme() string {
	return "etcd"
}

func (r Resolver) sync() error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(r.conf.RWTimeout)*time.Second)
	defer cancelFunc()
	// 返回前缀相同的多个
	// user/v1/xxx:11
	// user/v1/xxx:22
	get, err := r.etcdClient.Get(ctx, r.key, clientv3.WithPrefix())
	if err != nil {
		logs.Error("grpc client get etcd failed, name = %s, err:%v", r.key, err)
		return err
	}

	flag := true

	r.srcAddrList = []resolver.Address{}
	for _, v := range get.Kvs {
		server, err := ParseValue(v.Value)
		if err != nil {
			logs.Warn("grpc client parse etcd value failed, name = %s, err:%v", r.key, err)
			continue
		}
		flag = false
		// 通知 grpc
		r.srcAddrList = append(r.srcAddrList, resolver.Address{
			Addr:       server.Addr,
			Attributes: attributes.New("weight", server.Weight),
		})
	}
	if len(r.srcAddrList) == 0 {
		logs.Error("grpc client no service find")
		return nil
	}
	if flag {
		logs.Error("grpc client parse etcd value failed, name = %s", r.key)
		return err
	}
	err = r.cc.UpdateState(resolver.State{
		Addresses: r.srcAddrList,
	})
	if err != nil {
		logs.Error("grpc client update failed, name = %s, err:%v", r.key, err)
		return err
	}
	return nil
}

func (r Resolver) watch() {
	// 定时 周期同步数据
	// 监听节点的事件, 触发不同的操作
	// 监听close 事件, 关闭etcd
	r.watchCh = r.etcdClient.Watch(context.Background(), r.key, clientv3.WithPrefix())
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-r.closeCh:
			r.close()
		case res, ok := <-r.watchCh:
			if ok {
				r.update(res.Events)
			}
		case <-ticker.C:
			if err := r.sync(); err != nil {
				logs.Error("watch sync failed, err%v", err)
			}
		}
	}
}

func (r Resolver) update(events []*clientv3.Event) {
	for _, ev := range events {
		switch ev.Type {
		case clientv3.EventTypePut:
			// put key/value
			server, err := ParseValue(ev.Kv.Value)
			if err != nil {
				logs.Error("grpc client update value put parse failed, name = %s, err:%v", r.key, err)
			}
			addr := resolver.Address{
				Addr:       server.Addr,
				Attributes: attributes.New("weight", server.Weight),
			}
			if !Exist(r.srcAddrList, addr) {
				r.srcAddrList = append(r.srcAddrList, addr)
			}
			err = r.cc.UpdateState(resolver.State{
				Addresses: r.srcAddrList,
			})
			if err != nil {
				logs.Error("grpc client update value put failed, name = %s, err:%v", r.key, err)
			}
		case clientv3.EventTypeDelete:
			// 接收到 delete 操作, 删除 r.srvAddrList 其中匹配
			server, err := ParseKey(string(ev.Kv.Key))
			if err != nil {
				logs.Error("grpc client update value delete parse failed , name = %s, err:%v", r.key, err)
			}
			addr := resolver.Address{Addr: server.Addr}
			// 删除
			if list, ok := Remove(r.srcAddrList, addr); ok {
				r.srcAddrList = list
				err = r.cc.UpdateState(resolver.State{
					Addresses: r.srcAddrList,
				})
				if err != nil {
					logs.Error("grpc client update value delete failed, name = %s, err:%v", r.key, err)
				}
			}

		}
	}
}

func (r Resolver) close() {
	if r.etcdClient != nil {
		err := r.etcdClient.Close()
		if err != nil {
			logs.Error("close etcd err:%v ", err)
		}
	}
}

func Exist(list []resolver.Address, addr resolver.Address) bool {
	for i := range list {
		if list[i].Addr == addr.Addr {
			return true
		}
	}
	return false
}

func Remove(list []resolver.Address, addr resolver.Address) ([]resolver.Address, bool) {
	for i := range list {
		if list[i].Addr == addr.Addr {
			list[i] = list[len(list)-1]
			return list[:len(list)-1], true
		}
	}
	return nil, false
}

func NewResolver(conf config.EtcdConf) *Resolver {
	return &Resolver{
		conf:        conf,
		dialTimeout: conf.DialTimeout,
	}
}
