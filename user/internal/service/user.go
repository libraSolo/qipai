package service

import (
	"common/biz"
	"common/logs"
	"context"
	"core/dao"
	"core/models/entity"
	"core/models/requests"
	"core/repo"
	"framework/errorCode"
	"time"
	"user/pb"
)

// 创建账号

type AccountService struct {
	accountDao *dao.AccountDao
	redisDao   *dao.RedisDao
	pb.UnimplementedUserServiceServer
}

func NewAccountService(manager *repo.Manager) *AccountService {
	return &AccountService{
		accountDao: dao.NewAccountDao(manager),
		redisDao:   dao.NewRedisDao(manager),
	}
}

func (s *AccountService) Register(ctx context.Context, req *pb.RegisterParams) (*pb.RegisterResponse, error) {
	if req.LoginPlatform == requests.WeiXin {
		account, err := s.wxRegister(req)
		if err != nil {
			return &pb.RegisterResponse{}, errorCode.GrpcError(err)
		}
		return &pb.RegisterResponse{
			Uid: account.Uid,
		}, nil
	}

	logs.Info("register server called...")
	return &pb.RegisterResponse{}, nil
}

func (s *AccountService) wxRegister(req *pb.RegisterParams) (*entity.Account, *errorCode.Error) {
	// 1.封装一个 account 结构, 操作数据库 mongo 分布式id objectID
	account := &entity.Account{
		WxAccount:  req.Account,
		CreateTime: time.Now(),
	}
	exists, err := s.accountDao.Exists(context.TODO(), account)
	if err != nil {
		logs.Error("account register redis error err:%v", err)
		return account, biz.SqlError
	}
	if exists {
		logs.Error("account register redis error err:%v", err)
		return account, biz.InvalidUsers
	}
	// 2.生成数字作为用户的唯一id redis 自增
	uid, err := s.redisDao.NextAccountID()
	if err != nil {
		logs.Error("account register redis error err:%v", err)
		return account, biz.SqlError
	}
	account.Uid = uid

	err = s.accountDao.SaveAccount(context.TODO(), account)
	if err != nil {
		logs.Error("account save error err:%v", err)
		return account, biz.SqlError
	}

	return account, nil
}
