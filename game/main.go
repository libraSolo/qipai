package main

import (
	"common/config"
	"common/metrics"
	"context"
	"fmt"
	"framework/game"
	"game/app"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "game",
	Short: "game 游戏逻辑服",
	Long:  `game 游戏逻辑服`,
	Run: func(cmd *cobra.Command, args []string) {
	},
	PostRun: func(cmd *cobra.Command, args []string) {
	},
}

// var configFile = flag.String("config", "application.yml", "config file")
var (
	configFile    string
	gameConfigDir string
	serverId      string
)

func init() {
	rootCmd.Flags().StringVar(&configFile, "config", "application.yml", "app config yml file")
	rootCmd.Flags().StringVar(&gameConfigDir, "gameDir", "../config", "game config dir")
	rootCmd.Flags().StringVar(&serverId, "serverId", "", "app server id， required")
	_ = rootCmd.MarkFlagRequired("serverId")
}

func main() {
	// 加載配置
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	//flag.Parse()
	config.InitConfig(configFile)
	game.InitConfig(gameConfigDir)
	// 啓動監控
	go func() {
		err := metrics.Serve(fmt.Sprintf("0.0.0.0:%d", config.Conf.MetricPort))
		if err != nil {
			panic(err)
		}
	}()
	// 啓動 nats服务, 并进行订阅
	err := app.Run(context.Background(), serverId)
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
}
