package database

import (
	"common/config"
	"common/logs"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

type MongoManager struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func NewMongo() *MongoManager {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	clientOptions := options.Client().ApplyURI(config.Conf.Database.MongoConf.Url)
	clientOptions.SetAuth(options.Credential{
		Username: config.Conf.Database.MongoConf.UserName,
		Password: config.Conf.Database.MongoConf.Password,
	})

	// 设置连接池
	clientOptions.SetMinPoolSize(uint64(config.Conf.Database.MongoConf.MinPoolSize))
	clientOptions.SetMaxPoolSize(uint64(config.Conf.Database.MongoConf.MaxPoolSize))

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logs.Fatal("mongo connect error: %v", err)
		return nil
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		logs.Error("mongo ping error: %v", err)
		return nil
	}
	m := &MongoManager{
		Client: client,
	}
	m.DB = m.Client.Database(config.Conf.Database.MongoConf.Db)
	return m
}

func (m *MongoManager) Close() {
	err := m.Client.Disconnect(context.TODO())
	if err != nil {
		logs.Fatal("mongo close error: %v", err)
	}
}
