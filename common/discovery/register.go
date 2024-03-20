package discovery

import (
	"common/config"
	"common/logs"
	"context"
	"encoding/json"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

// Register register grpc 服务注册到 etcd
type Register struct {
	etcdClient  *clientv3.Client                        // etcd 连接
	leaseID     clientv3.LeaseID                        // 租约 ID
	dialTimeout int                                     // 超时时间
	ttl         int                                     // 续约时间
	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse // 心跳
	info        Server                                  // 注册信息
	closeCh     chan struct{}                           // 关闭标识
}

func NewRegister() *Register {
	return &Register{dialTimeout: 3}
}

func (r *Register) Close() {
	r.closeCh <- struct{}{}
}

func (r *Register) Register(config config.EtcdConf) (err error) {
	// 注册信息
	info := Server{
		Name:    config.Register.Name,
		Addr:    config.Register.Addr,
		Weight:  config.Register.Weight,
		Version: config.Register.Version,
		Ttl:     config.Register.Ttl,
	}

	// 建立 etcd 的连接
	r.etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   config.Addrs,
		DialTimeout: time.Duration(r.dialTimeout) * time.Second,
	})
	if err != nil {
		return err
	}

	r.info = info
	if err = r.register(); err != nil {
		return err
	}
	r.closeCh = make(chan struct{})
	// 根据心跳做操作
	go r.watcher()
	return
}

func (r *Register) register() error {
	// 1.创建租约
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(r.dialTimeout))
	defer cancel()
	var err error
	if err = r.createLease(ctx, r.info.Ttl); err != nil {
		return err
	}
	// 2.心跳监测
	if r.keepAliveCh, err = r.keepAlive(); err != nil {
		return err
	}
	// 3.绑定租约
	data, _ := json.Marshal(r.info)
	return r.bindLease(ctx, r.info.BuildRegisterKey(), string(data))
}

// 创建租约 ttl 秒
func (r *Register) createLease(ctx context.Context, ttl int64) error {
	grant, err := r.etcdClient.Grant(ctx, ttl)
	if err != nil {
		logs.Error("error creating lease failed: %v", err)
		return err
	}
	r.leaseID = grant.ID
	return nil
}

// 绑定租约
func (r *Register) bindLease(ctx context.Context, key, value string) error {
	// put
	_, err := r.etcdClient.Put(ctx, key, value, clientv3.WithLease(r.leaseID))
	if err != nil {
		logs.Error("error blind lease failed: %v", err)
		return err
	}
	logs.Info("register service succeeded key: %v", key)
	return nil
}

// 心跳监测
func (r *Register) keepAlive() (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	keepAliveResponses, err := r.etcdClient.KeepAlive(context.Background(), r.leaseID)
	if err != nil {
		logs.Error("error blind lease failed: %v", err)
		return keepAliveResponses, err
	}
	return keepAliveResponses, nil
}

// 续约 || 新注册
func (r *Register) watcher() {
	// 到期， 检查是否需要注册
	ticker := time.NewTicker(time.Duration(r.info.Ttl) * time.Second)
	for {
		select {
		case <-r.closeCh:
			if err := r.unregister(); err != nil {
				logs.Error("close and unregister failed, err%v", err)
			}
			// 租约撤销
			if _, err := r.etcdClient.Revoke(context.Background(), r.leaseID); err != nil {
				logs.Error("close and Revoke lease failed, err:%v", err)
			}
			if r.etcdClient != nil {
				r.etcdClient.Close()
			}
			logs.Info("unregister etcd...")
		case res := <-r.keepAliveCh:
			//logs.Info("keep alive %v", res)
			if res == nil {
				if err := r.register(); err != nil {
					logs.Error("keepAliveCh register failed, err:%v", err)
				}
				logs.Info("etcd 重新注册成功")
			}
		case <-ticker.C:
			if r.keepAliveCh == nil {
				if err := r.register(); err != nil {
					logs.Error("ticker register failed, err:%v", err)
				}
			}
		}
	}
}

func (r *Register) unregister() error {
	_, err := r.etcdClient.Delete(context.Background(), r.info.BuildRegisterKey())
	if err != nil {
		return err
	}
	return nil
}
