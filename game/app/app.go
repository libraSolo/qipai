package app

import (
	"common/config"
	"common/logs"
	"context"
	"core/repo"
	"fmt"
	"framework/node"
	"game/route"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run(ctx context.Context, serverId string) error {
	// log
	logs.InitLog(config.Conf.AppName)
	exit := func() {}
	manager := repo.NewManager()
	go func() {
		// 1.wsManager 2.natsClient
		n := node.Default()
		exit = n.Close
		n.RegisterHandler(route.Register(manager))
		n.Run(serverId)
	}()

	stop := func() {
		exit()
		// 等待时间
		time.Sleep(5 * time.Second)
		fmt.Println("stop connector finished")
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
				logs.Info("connector app quit")
				return nil
			case syscall.SIGHUP:
				stop()
				logs.Info("hang up, connector app quit")
			default:
				return nil
			}
		}
	}
}
