package app

import (
	"common/config"
	"common/discovery"
	"common/logs"
	"context"
	"core/repo"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user/internal/service"
	"user/pb"
)

// Run 启动 grpc http log db
func Run(ctx context.Context) error {
	// log
	logs.InitLog(config.Conf.AppName)
	// etcd 注册中心 grpc服务注册到 etcd 中
	register := discovery.NewRegister()
	// 启动 grpc
	server := grpc.NewServer()
	// todo 初始化 数据库管理
	manager := repo.NewManager()

	go func() {
		listen, err := net.Listen("tcp", config.Conf.Grpc.Addr)
		if err != nil {
			logs.Fatal("user grpc run listen err: %v", err)
		}
		// 注册 grpc service 到 tecd 中
		err = register.Register(config.Conf.Etcd)
		if err != nil {
			logs.Fatal("user grpc register err: %v", err)
		}

		pb.RegisterUserServiceServer(server, service.NewAccountService(manager))
		// 阻塞操作
		err = server.Serve(listen)
		if err != nil {
			logs.Fatal("user grpc server listen err: %v", err)
		}
	}()

	stop := func() {
		server.Stop()
		register.Close()
		manager.Close()
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
