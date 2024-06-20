package sz

import (
	"common/utils"
	"sort"
	"sync"
)

type Logic struct {
	sync.RWMutex
	cards []int
}

func NewLogic() *Logic {
	return &Logic{
		cards: make([]int, 0),
	}
}

// 黑红梅方
func (l *Logic) washCards() {
	l.cards = []int{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d,
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d,
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d,
		0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d,
	}
	for i, v := range l.cards {
		random := utils.Rand(len(l.cards))
		l.cards[i] = l.cards[random]
		l.cards[random] = v
	}
}

// 获取三张手牌
func (l *Logic) getCards() []int {
	cards := make([]int, 3)
	l.RLock()
	defer l.RUnlock()
	for i := 0; i < 3; i++ {
		if len(cards) == 0 {
			break
		}
		card := l.cards[len(l.cards)-1]
		l.cards = l.cards[:len(l.cards)-1]
		cards[i] = card
	}
	return cards
}

// CompareCards 0:平 大于0:胜 小于0:负
func (l *Logic) CompareCards(cur []int, other []int) int {
	// 获取牌型
	curType := l.getCardType(cur)
	otherType := l.getCardType(other)
	if curType != otherType {
		return int(curType - otherType)
	}
	// 同牌型比大小
	if curType == DuiZi {
		duiZi, one := l.getDuiZi(cur)
		duiZiOther, oneOther := l.getDuiZi(other)
		if duiZi != duiZiOther {
			return duiZi - duiZiOther
		}
		return one - oneOther
	}
	curValue := l.getSortCards(cur)
	otherValue := l.getSortCards(other)
	if otherValue[2] != curValue[2] {
		return otherValue[2] - curValue[2]
	}
	if otherValue[1] != curValue[1] {
		return otherValue[1] - curValue[1]
	}
	if otherValue[0] != curValue[0] {
		return otherValue[0] - curValue[0]
	}

	return 0
}

func (l *Logic) getCardType(cards []int) CardsType {
	values := l.getSortCards(cards)
	oneV := values[0]
	twoV := values[1]
	threeV := values[2]
	// 1.豹子 牌面值相等(因为颜色不同,先获得数值)
	if oneV == twoV && twoV == threeV {
		return BaoZi
	}
	// 2. 金花 颜色相同 顺子
	color := false
	oneColor := l.getCardColor(cards[0])
	twoColor := l.getCardColor(cards[1])
	threeColor := l.getCardColor(cards[2])
	if oneColor == twoColor && twoColor == threeColor {
		color = true
	}
	// 3. 顺子
	shunZi := false
	if (oneV+1 == twoV && twoV+1 == threeV) || (oneV == 2 && twoV == 3 && threeV == 14) {
		shunZi = true
	}
	if color && shunZi {
		return ShunJin
	}
	if color {
		return JinHua
	}
	if shunZi {
		return ShunZi
	}
	// 4. 对子
	if oneV == twoV || twoV == threeV {
		return DuiZi
	}
	return DanZhang
}

func (l *Logic) getSortCards(cards []int) []int {
	// 映射 1-13 -> 2-14 将 1变成14
	v := make([]int, len(cards))
	for i, card := range cards {
		value := l.getCardNumber(card)
		if value == 1 {
			v[i] = 14
		}
		v[i] = value + 1
	}
	sort.Ints(v)
	return v
}

func (l *Logic) getCardNumber(card int) int {
	return card & 0x0f
}

func (l *Logic) getCardColor(card int) string {
	colors := []string{"黑", "红", "梅", "方"}
	if card < 0x01 || card > 0x3d {
		return ""
	}
	return colors[card/0x10]
}

func (l *Logic) getDuiZi(cards []int) (int, int) {
	values := l.getSortCards(cards)
	oneV := values[0]
	twoV := values[1]
	threeV := values[2]
	if oneV == twoV {
		// AAB
		return twoV, threeV
	}
	// ABB
	return twoV, oneV
}
