package database

import (
	"common/config"
	"common/logs"
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisManager struct {
	Client        *redis.Client        // 单机
	ClusterClient *redis.ClusterClient // 集群
}

func NewRedisManager() *RedisManager {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	var client *redis.Client
	var ClusterClient *redis.ClusterClient
	redisConfig := config.Conf.Database.RedisConf
	// 判断模式
	if len(redisConfig.ClusterAddrs) == 0 {
		// 单机
		client = redis.NewClient(&redis.Options{
			Addr:         redisConfig.Addr,
			Password:     redisConfig.Password,
			PoolSize:     redisConfig.PoolSize,
			MinIdleConns: redisConfig.MinIdleConns,
		})
		if err := client.Ping(ctx).Err(); err != nil {
			logs.Fatal("redis client ping error: %v", err)
			return nil
		}
	} else {
		// 集群
		ClusterClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        redisConfig.ClusterAddrs,
			Password:     redisConfig.Password,
			PoolSize:     redisConfig.PoolSize,
			MinIdleConns: redisConfig.MinIdleConns,
		})
		if err := ClusterClient.Ping(ctx).Err(); err != nil {
			logs.Fatal("redis client ping error: %v", err)
			return nil
		}
	}
	return &RedisManager{
		Client:        client,
		ClusterClient: ClusterClient,
	}
}

func (r *RedisManager) Close() {
	if r.Client != nil {
		if err := r.Client.Close(); err != nil {
			logs.Error("redis client ping error: %v", err)
		}
		return
	}

	if r.ClusterClient != nil {
		if err := r.ClusterClient.Close(); err != nil {
			logs.Error("redis clusterClient ping error: %v", err)
		}
		return
	}
}
