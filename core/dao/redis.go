package dao

import (
	"context"
	"core/repo"
	"fmt"
)

const Prefix = "QiPai"
const AccountIdRedisKey = "AccountId"
const AccountIdBegin = 10000

type RedisDao struct {
	repo *repo.Manager
}

func (d *RedisDao) NextAccountID() (string, error) {
	// 自增
	return d.incr(Prefix + ":" + AccountIdRedisKey)
}

func (d *RedisDao) incr(key string) (string, error) {
	// 判断 key 是否存在, 不存在 set 存在就自增
	var result int64
	var err error
	if d.repo.Redis.Client != nil {
		result, err = d.repo.Redis.Client.Exists(context.TODO(), key).Result()
	} else {
		result, err = d.repo.Redis.ClusterClient.Exists(context.TODO(), key).Result()
	}
	if err != nil {
		return "", err
	}
	if 0 == result {
		// 不存在
		if d.repo.Redis.Client != nil {
			err = d.repo.Redis.Client.Set(context.TODO(), key, AccountIdBegin, 0).Err()
		} else {
			err = d.repo.Redis.ClusterClient.Set(context.TODO(), key, AccountIdBegin, 0).Err()
		}
		if err != nil {
			return "", err
		}
	}
	var id int64
	if d.repo.Redis.Client != nil {
		id, err = d.repo.Redis.Client.Incr(context.TODO(), key).Result()
	} else {
		id, err = d.repo.Redis.ClusterClient.Incr(context.TODO(), key).Result()
	}
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", id), nil
}

func NewRedisDao(m *repo.Manager) *RedisDao {
	return &RedisDao{
		repo: m,
	}
}
