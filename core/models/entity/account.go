package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Account struct {
	Id           primitive.ObjectID `bson:"_id,omitempty"`
	Uid          string             `bson:"uid"`
	Account      string             `bson:"account" `
	Password     string             `bson:"password"`
	PhoneAccount string             `bson:"phoneAccount"`
	WxAccount    string             `bson:"wxAccount"`
	CreateTime   time.Time          `bson:"createTime"`
}
