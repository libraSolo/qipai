package service

import (
	"common/logs"
	"common/utils"
	"connector/models/request"
	"context"
	"core/dao"
	"core/models/entity"
	"core/repo"
	"fmt"
	"framework/game"
	"time"
)

type UserService struct {
	userDao *dao.UserDao
}

func NewUserService(r *repo.Manager) *UserService {
	return &UserService{
		userDao: dao.NewUserDao(r),
	}
}

func (s UserService) FindUserByUid(ctx context.Context, uid string, info request.UserInfo) (*entity.User, error) {
	// 查询 mongo 存在则返回，没有则新增
	user, err := s.userDao.FindUserByUid(ctx, uid)
	if err != nil {
		logs.Error("find user mongo err:%v", err)
		return nil, err
	}
	// 创建角色
	if user == nil {
		user = &entity.User{}
		user.Uid = uid
		user.Gold = int64(game.Conf.GameConfig["startgold"]["value"].(float64))
		user.Avatar = utils.Default(info.Avatar, "Default")
		user.Nickname = utils.Default(info.Nickname, fmt.Sprintf("%s%s", "用户", uid))
		user.Sex = info.Sex
		user.CreateTime = time.Now().Unix()
		user.LastLoginTime = time.Now().Unix()
		err = s.userDao.Insert(context.TODO(), user)
		if err != nil {
			logs.Error("insert user mongo err:%v", err)
			return nil, err
		}
	}
	return user, nil
}
