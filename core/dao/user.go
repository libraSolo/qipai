package dao

import (
	"context"
	"core/models/entity"
	"core/repo"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserDao struct {
	repo *repo.Manager
}

func NewUserDao(m *repo.Manager) *UserDao {
	return &UserDao{
		repo: m,
	}
}

func (d UserDao) FindUserByUid(ctx context.Context, uid string) (*entity.User, error) {
	db := d.repo.Mongo.DB.Collection("user")
	result := db.FindOne(ctx, bson.D{{"uid", uid}})
	user := new(entity.User)
	err := result.Decode(user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (d UserDao) Insert(ctx context.Context, user *entity.User) error {
	db := d.repo.Mongo.DB.Collection("user")
	_, err := db.InsertOne(ctx, user)
	return err
}
