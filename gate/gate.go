package gate

import (
	"common/config"
	"common/metrics"
	"context"
	"flag"
	"fmt"
	"gate/app"
	"log"
	"os"
)

var configFile = flag.String("config", "application.yml", "config file")

func main() {
	// 加載配置
	flag.Parse()
	config.InitConfig(*configFile)
	// 啓動監控
	go func() {
		err := metrics.Serve(fmt.Sprintf("0.0.0.0:%d", config.Conf.MetricPort))
		if err != nil {
			panic(err)
		}
	}()
	// 啓動 grpc 服務器
	err := app.Run(context.Background())
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
}
