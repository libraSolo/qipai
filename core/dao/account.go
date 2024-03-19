package dao

import (
	"context"
	"core/models/entity"
	"core/repo"
)

type AccountDao struct {
	repo *repo.Manager
}

func NewAccountDao(m *repo.Manager) *AccountDao {
	return &AccountDao{
		repo: m,
	}
}

func (d AccountDao) SaveAccount(ctx context.Context, account *entity.Account) error {
	table := d.repo.Mongo.DB.Collection("account")
	_, err := table.InsertOne(ctx, account)
	if err != nil {
		return err
	}
	return nil
}

func (d AccountDao) Exists(ctx context.Context, account *entity.Account) (bool, error) {
	table := d.repo.Mongo.DB.Collection("account")
	_, err := table.Find(ctx, account)
	if err != nil {
		return false, err
	}
	return true, nil
}
