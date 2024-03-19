package game

import (
	"common/logs"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"os"
	"path"
)

var Conf *Config

const (
	gameConfig = "gameConfig.json"
	servers    = "servers.json"
)

type Config struct {
	GameConfig  map[string]GameConfigValue `json:"gameConfig"`
	ServersConf ServersConf                `json:"serversConf"`
}
type ServersConf struct {
	Nats       NatsConfig         `json:"nats"`
	Connector  []*ConnectorConfig `json:"connector"`
	Servers    []*ServersConfig   `json:"servers"`
	TypeServer map[string][]*ServersConfig
}

type ServersConfig struct {
	ID               string `json:"id"`
	ServerType       string `json:"serverType"`
	HandleTimeOut    int    `json:"handleTimeOut"`
	RPCTimeOut       int    `json:"rpcTimeOut"`
	MaxRunRoutineNum int    `json:"maxRunRoutineNum"`
}

type ConnectorConfig struct {
	ID         string `json:"id"`
	Host       string `json:"host"`
	ClientPort int    `json:"clientPort"`
	Frontend   bool   `json:"frontend"`
	ServerType string `json:"serverType"`
}
type NatsConfig struct {
	Url string `json:"url"`
}

type GameConfigValue map[string]interface{}

func InitConfig(configDir string) {
	Conf = new(Config)
	dir, err := os.ReadDir(configDir)
	if err != nil {
		logs.Fatal("read config dir err:%v", err)
	}
	for _, v := range dir {
		configFile := path.Join(configDir, v.Name())
		if v.Name() == gameConfig {
			readGameConfig(configFile)
		} else if v.Name() == servers {
			readServersConfig(configFile)
		}
	}
}

func readServersConfig(configFile string) {
	var serversConfig ServersConf
	v := viper.New()
	v.SetConfigFile(configFile)
	// 修改后重新加载
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		log.Println("servers config changed")
		// 解析
		err := v.Unmarshal(&serversConfig)
		if err != nil {
			panic(fmt.Errorf("servers config changed unmarshal config %s: %v", configFile, err))
		}
		Conf.ServersConf = serversConfig
	})
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("servers config reading config:%s: %v", configFile, err))
	}

	// 解析
	err = v.Unmarshal(&serversConfig)
	if err != nil {
		panic(fmt.Errorf("servers config unmarshal config %s: %v", configFile, err))
	}
	Conf.ServersConf = serversConfig
	typeServersConfig()
}

func typeServersConfig() {
	if len(Conf.ServersConf.Servers) > 0 {
		if Conf.ServersConf.TypeServer == nil {
			Conf.ServersConf.TypeServer = make(map[string][]*ServersConfig)
		}
		for _, v := range Conf.ServersConf.Servers {
			if Conf.ServersConf.TypeServer[v.ServerType] == nil {
				Conf.ServersConf.TypeServer[v.ServerType] = make([]*ServersConfig, 0)
			}
			Conf.ServersConf.TypeServer[v.ServerType] = append(Conf.ServersConf.TypeServer[v.ServerType], v)
		}
	}
}

func readGameConfig(configFile string) {
	var gameConfig = make(map[string]GameConfigValue)
	v := viper.New()
	v.SetConfigFile(configFile)
	// 修改后重新加载
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		log.Println("game config changed")
		// 解析
		err := v.Unmarshal(&gameConfig)
		if err != nil {
			panic(fmt.Errorf("game config changed unmarshal config %s: %v", configFile, err))
		}
		Conf.GameConfig = gameConfig
	})
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("game config reading config:%s: %v", configFile, err))
	}

	// 解析
	err = v.Unmarshal(&gameConfig)
	if err != nil {
		panic(fmt.Errorf("game config unmarshal config %s: %v", configFile, err))
	}
	Conf.GameConfig = gameConfig
}

func (c *Config) GetConnector(serverId string) *ConnectorConfig {
	for _, v := range c.ServersConf.Connector {
		if v.ID == serverId {
			return v
		}
	}
	return nil
}

func (c *Config) GetConnectorByServerType(serverType string) *ConnectorConfig {
	for _, config := range c.ServersConf.Connector {
		if config.ServerType == serverType {
			return config
		}
	}
	return nil
}

func (c *Config) GetFrontGameConfig() map[string]any {
	result := make(map[string]any)
	for s, v := range c.GameConfig {
		value, ok := v["value"]
		backend := false
		_, exist := v["backend"]
		if exist {
			backend = v["backend"].(bool)
		}
		if ok && !backend {
			result[s] = value
		}
	}
	return result
}
