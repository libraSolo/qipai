package mj

import (
	"common/utils"
	"sync"
)

type CardID int

type CardIDs []CardID

func (c CardIDs) Len() int {
	return len(c)
}
func (c CardIDs) Less(i, j int) bool {
	return c[i] < c[j]
}
func (c CardIDs) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

const (
	Wan1  CardID = 1
	Wan2  CardID = 2
	Wan3  CardID = 3
	Wan4  CardID = 4
	Wan5  CardID = 5
	Wan6  CardID = 6
	Wan7  CardID = 7
	Wan8  CardID = 8
	Wan9  CardID = 9
	Tong1 CardID = 11
	Tong2 CardID = 12
	Tong3 CardID = 13
	Tong4 CardID = 14
	Tong5 CardID = 15
	Tong6 CardID = 16
	Tong7 CardID = 17
	Tong8 CardID = 18
	Tong9 CardID = 19
	Tiao1 CardID = 21
	Tiao2 CardID = 22
	Tiao3 CardID = 23
	Tiao4 CardID = 24
	Tiao5 CardID = 25
	Tiao6 CardID = 26
	Tiao7 CardID = 27
	Tiao8 CardID = 28
	Tiao9 CardID = 29
	Dong  CardID = 31
	Nan   CardID = 32
	Xi    CardID = 33
	Bei   CardID = 34
	Zhong CardID = 35
)

type Logic struct {
	sync.RWMutex
	cards    []CardID
	gameType GameType
	qidui    bool // 小七对
}

func NewLogic(gameType GameType, qidui bool) *Logic {
	return &Logic{
		gameType: gameType,
		qidui:    qidui,
	}
}

func (l *Logic) washCards() {
	l.Lock()
	defer l.Unlock()

	l.cards = []CardID{
		Tong1, Tong2, Tong3, Tong4, Tong5, Tong6, Tong7, Tong8, Tong9,
		Tong1, Tong2, Tong3, Tong4, Tong5, Tong6, Tong7, Tong8, Tong9,
		Tong1, Tong2, Tong3, Tong4, Tong5, Tong6, Tong7, Tong8, Tong9,
		Tong1, Tong2, Tong3, Tong4, Tong5, Tong6, Tong7, Tong8, Tong9,
		Tiao1, Tiao2, Tiao3, Tiao4, Tiao5, Tiao6, Tiao7, Tiao8, Tiao9,
		Tiao1, Tiao2, Tiao3, Tiao4, Tiao5, Tiao6, Tiao7, Tiao8, Tiao9,
		Tiao1, Tiao2, Tiao3, Tiao4, Tiao5, Tiao6, Tiao7, Tiao8, Tiao9,
		Tiao1, Tiao2, Tiao3, Tiao4, Tiao5, Tiao6, Tiao7, Tiao8, Tiao9,
		Wan1, Wan2, Wan3, Wan4, Wan5, Wan6, Wan7, Wan8, Wan9,
		Wan1, Wan2, Wan3, Wan4, Wan5, Wan6, Wan7, Wan8, Wan9,
		Wan1, Wan2, Wan3, Wan4, Wan5, Wan6, Wan7, Wan8, Wan9,
		Wan1, Wan2, Wan3, Wan4, Wan5, Wan6, Wan7, Wan8, Wan9,
		Zhong, Zhong, Zhong, Zhong,
	}

	if l.gameType == HongZhong8 {
		l.cards = append(l.cards, Zhong, Zhong, Zhong, Zhong)
	}

	// 洗牌
	for i := 0; i < 300; i++ {
		index := i % len(l.cards) // 防止越界
		random := utils.Rand(len(l.cards))
		l.cards[index], l.cards[random] = l.cards[random], l.cards[index]
	}
}

func (l *Logic) getCards(num int) []CardID {
	// 发牌之后 牌没了
	l.Lock()
	defer l.Unlock()

	if len(l.cards) < num {
		return nil
	}

	cards := l.cards[:num]
	l.cards = l.cards[num:]

	return cards
}
