package connector

import (
	"common/logs"
	"fmt"
	"framework/game"
	"framework/net"
	"framework/remote"
)

type Connector struct {
	isRunning bool
	wsManager *net.Manager
	handlers  net.LogicHandler
	remoteCli remote.Client
}

func Default() *Connector {
	return &Connector{
		wsManager: &net.Manager{},
		handlers:  make(net.LogicHandler),
	}
}

func (c *Connector) Run(serverId string) {
	if !c.isRunning {
		// 启动 ws 和 nats
		c.wsManager = net.NewManager()
		c.remoteCli = remote.NewNatsClient(serverId, c.wsManager.RemoteReadChan)
		c.remoteCli.Run()
		c.wsManager.ConnectorHandlers = c.handlers
		c.wsManager.RemoteCli = c.remoteCli
		c.serve(serverId)
	}
}

func (c *Connector) Close() {
	if c.isRunning {
		// 关闭 ws 和 nats
		c.wsManager.Close()
	}
}

func (c *Connector) serve(serverId string) {
	logs.Info("run connector :%v", serverId)
	// 地址
	connectorConfig := game.Conf.GetConnector(serverId)
	if connectorConfig == nil {
		logs.Fatal("get connector config failed")
	}
	addr := fmt.Sprintf("%s:%d", connectorConfig.Host, connectorConfig.ClientPort)
	c.wsManager.ServerId = serverId
	c.isRunning = true
	c.wsManager.Run(addr)
}

func (c *Connector) RegisterHandler(handlers net.LogicHandler) {
	c.handlers = handlers
}
