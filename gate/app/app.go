package app

import (
	"common/config"
	"common/logs"
	"context"
	"fmt"
	"gate/router"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Run 启动 grpc http log db
func Run(ctx context.Context) error {
	// log
	logs.InitLog(config.Conf.AppName)

	go func() {
		// gin 启动 注册一个路由
		r := router.RegisterRouter()

		// http 接口
		err := r.Run(fmt.Sprintf(":%v", config.Conf.HttpPort))
		if err != nil {
			logs.Error("gate gin run err:%v", err)
			return
		}
	}()

	stop := func() {
		// 等待时间
		time.Sleep(5 * time.Second)
		fmt.Println("stop user finished")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGHUP)
	for {
		select {
		case <-ctx.Done():
			stop()
			return nil
		case s := <-c:
			switch s {
			case syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT:
				stop()
				logs.Info("user app quit")
				return nil
			case syscall.SIGHUP:
				stop()
				logs.Info("hang up, user app quit")
			default:
				return nil
			}
		}
	}
}
