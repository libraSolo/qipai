package remote

import (
	"common/logs"
	"framework/game"
	"github.com/nats-io/nats.go"
)

type NatsClient struct {
	serverId string
	conn     *nats.Conn
	readChan chan []byte
}

func NewNatsClient(serverId string, readChan chan []byte) *NatsClient {
	return &NatsClient{
		serverId: serverId,
		readChan: readChan,
	}
}

func (c *NatsClient) Run() error {
	var err error
	c.conn, err = nats.Connect(game.Conf.ServersConf.Nats.Url)
	if err != nil {
		logs.Error("Connect nats failed err:%v", err)
		return err
	}
	go c.sub()
	return nil
}

func (c *NatsClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *NatsClient) SendMsg(dst string, data []byte) error {
	if c.conn != nil {
		err := c.conn.Publish(dst, data)
		if err != nil {
			logs.Error("Publish nats failed err:%v", err)
			return err
		}
	}
	return nil
}

func (c *NatsClient) sub() {
	_, err := c.conn.Subscribe(c.serverId, func(msg *nats.Msg) {
		// 收到其他nats client 发送的消息
		c.readChan <- msg.Data
	})
	if err != nil {
		logs.Error("Subscribe nats failed err:%v", err)
	}
}
