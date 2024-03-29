package logic

import (
	"fmt"
	"game/component/room"
	"golang.org/x/exp/rand"
	"sync"
	"time"
)

var (
	baseUnionId int64 = 10000
)

type UnionManager struct {
	sync.RWMutex
	unionList map[int64]*Union
}

func NewUnionManager() *UnionManager {
	list := make(map[int64]*Union)
	u := &UnionManager{
		unionList: list,
	}
	// 创建默认为 1 的联盟
	u.CreateUnionById(1)
	return u
}

func (m *UnionManager) GetUnion(id int64) *Union {
	m.RLock()
	defer m.RUnlock()
	return m.unionList[id]
}

func (m *UnionManager) CreateUnionById(id int64) *Union {
	m.Lock()
	defer m.Unlock()
	union := NewUnion(m)
	m.unionList[id] = union
	return union
}

func (m *UnionManager) CreateUnion() *Union {
	m.Lock()
	defer m.Unlock()
	u := m.GetUnion(baseUnionId)
	if u != nil {
		return nil
	}
	union := NewUnion(m)
	m.unionList[baseUnionId] = union
	baseUnionId++
	return union
}

func (m *UnionManager) CreateRoomId() string {
	// 随机数
	// TODO: redis 创建随机Id
	roomId := m.genRoomId()
	for _, v := range m.unionList {
		_, ok := v.RoomList[roomId]
		if ok {
			return m.CreateRoomId()
		}
	}
	return roomId
}

func (m *UnionManager) genRoomId() string {
	rand.Seed(uint64(time.Now().UnixNano()))
	// 房间号6位数
	roomIdInt := rand.Int63n(899999) + 10000
	return fmt.Sprintf("%d", roomIdInt)
}

func (m *UnionManager) GetRoomById(s string) *room.Room {
	for _, v := range m.unionList {
		r, ok := v.RoomList[s]
		if ok {
			return r
		}
	}
	return nil
}
