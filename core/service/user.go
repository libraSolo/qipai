package service

import (
	"common/logs"
	"common/utils"
	"connector/models/request"
	"context"
	"core/dao"
	"core/models/entity"
	"core/repo"
	"errors"
	"fmt"
	"framework/game"
	hallReq "hall/models/request"
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

func (s *UserService) FindUserByUid(ctx context.Context, uid string) (*entity.User, error) {
	// 查询 mongo 存在则返回，没有则新增
	user, err := s.userDao.FindUserByUid(ctx, uid)
	if err != nil {
		logs.Error("find user mongo err:%v", err)
		return nil, err
	}
	return user, nil
}

func (s *UserService) CreateUserByUid(ctx context.Context, uid string, info request.UserInfo) (*entity.User, error) {
	// 查询 mongo 存在则返回，没有则新增
	user, err := s.userDao.FindUserByUid(ctx, uid)
	if err != nil {
		logs.Error("find user mongo err:%v", err)
		return nil, err
	}
	// 创建角色
	if user != nil {
		logs.Error("create user exist")
		return user, errors.New("user exist")
	}
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
	return user, nil
}

func (s *UserService) UpdateUserAddressByUid(uid string, req hallReq.UpdateUserAddressReq) error {
	user := &entity.User{
		Uid:      uid,
		Address:  req.Address,
		Location: req.Location,
	}
	err := s.userDao.UpdateUserAddressByUid(context.TODO(), user)
	if err != nil {
		logs.Error("update user err:%v", err)
		return err
	}
	return nil
}
