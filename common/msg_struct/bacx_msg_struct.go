package msg_struct

// id =GameStatusChangeProtocol = 20103
type BacxChangeGameStatus struct {
	DeskNo     int `json:"desk_no"`     //桌号
	Set        int `json:"set"`         //场次
	Round      int `json:"round"`       //局号
	GameStatus int `json:"game_status"` //游戏状态
}

// 龙虎游戏的牌
type BacxAnswerCard struct {
	PlayerCard [3]Card `json:"pcard"` //庄的牌
	BankerCard [3]Card `json:"bcard"` //闲的牌
}

// 得到闲家点数
func (baAnswer *BacxAnswerCard) GetPlayerPoint() int {

	point := 0

	for i := 0; i < 3; i++ {
		if baAnswer.PlayerCard[i].Used == 0 {
			continue
		}

		//点数通过10的不处理
		if baAnswer.PlayerCard[i].Number >= 10 {
			continue
		}

		point += int(baAnswer.PlayerCard[i].Number)
		point = point % 10

	}

	return point

}

// 得到庄家点数
func (baAnswer *BacxAnswerCard) GetBankerPoint() int {

	point := 0

	for i := 0; i < 3; i++ {
		if baAnswer.BankerCard[i].Used == 0 {
			continue
		}

		//点数通过10的不处理
		if baAnswer.BankerCard[i].Number >= 10 {
			continue
		}

		point += int(baAnswer.BankerCard[i].Number)
		point = point % 10

	}

	return point
}

// 得到闲家牌数
func (baAnswer *BacxAnswerCard) GetPlayerNum() int {

	num := 0

	for i := 0; i < 3; i++ {
		if baAnswer.PlayerCard[i].Used == 0 {
			continue
		}
		num++
	}

	return num
}

// 得到庄家牌数
func (baAnswer *BacxAnswerCard) GetBankerNum() int {

	num := 0

	for i := 0; i < 3; i++ {
		if baAnswer.BankerCard[i].Used == 0 {
			continue
		}
		num++
	}

	return num
}

// 检测庄家是否刚好以6点胜
func (baAnswer *BacxAnswerCard) IsSuperSix() bool {

	PlayerPoint := baAnswer.GetPlayerPoint()
	BankerPoint := baAnswer.GetBankerPoint()

	if BankerPoint != 6 {
		return false
	}

	//庄家刚好以6点胜,才是supersix
	if PlayerPoint >= BankerPoint {
		return false
	}

	return true
}

// 检测是否完美对子(点数和色彩都一样)
func (baAnswer *BacxAnswerCard) IsPerfectPair() bool {

	//先检测闲
	if baAnswer.PlayerCard[0].Number == baAnswer.PlayerCard[1].Number {

		if baAnswer.PlayerCard[0].Color == baAnswer.PlayerCard[1].Color {
			return true
		}
	}
	//检测庄
	if baAnswer.BankerCard[0].Number == baAnswer.BankerCard[1].Number {

		if baAnswer.BankerCard[0].Color == baAnswer.BankerCard[1].Color {
			return true
		}

	}
	return false
}

// 检测是否任意对子(闲或庄前两张牌一样，色彩可以不一样)
func (baAnswer *BacxAnswerCard) IsAnyPair() bool {

	//先检测闲
	if baAnswer.PlayerCard[0].Number == baAnswer.PlayerCard[1].Number {
		return true
	}

	//检测庄
	if baAnswer.BankerCard[0].Number == baAnswer.BankerCard[1].Number {
		return true
	}

	return false
}

// 是否闲龙宝,返回的第二个差数，是相差的点数
func (baAnswer *BacxAnswerCard) IsPlayerBonus() bool {

	PlayerPoint := baAnswer.GetPlayerPoint()
	BankerPoint := baAnswer.GetBankerPoint()

	cardnum := baAnswer.GetPlayerNum()

	if PlayerPoint <= BankerPoint {
		return false
	}

	if PlayerPoint-BankerPoint <= 3 {
		return false
	}

	if cardnum == 2 {

		//若是牌例胜了
		if PlayerPoint == 8 || PlayerPoint == 9 {
			return true
		}

		//只有2张牌的检测是否7点以下
		if PlayerPoint <= 7 {
			return true
		}

		return false
	}

	//以下是三张牌的
	//winPoint := PlayerPoint - BankerPoint //相差的点数为4---9

	return true
}

// 是否庄龙宝,返回的第二个差数，是相差的点数
func (baAnswer *BacxAnswerCard) IsBankerBonus() bool {

	PlayerPoint := baAnswer.GetPlayerPoint()
	BankerPoint := baAnswer.GetBankerPoint()
	cardnum := baAnswer.GetBankerNum()

	if PlayerPoint >= BankerPoint {
		return false
	}

	//点数相差<=3,则不符合条件
	if BankerPoint-PlayerPoint <= 3 {
		return false
	}

	//若是只有2张牌
	if cardnum == 2 {

		//若是牌例胜了
		if BankerPoint == 8 || BankerPoint == 9 {
			return true
		}

		//只有2张牌的检测是否7点以下
		if BankerPoint <= 7 {
			return true
		}

		return false
	}

	//以下是三张牌的
	//winPoint := PlayerPoint - BankerPoint //相差的点数为4---9

	return true

}

// 检测是否大，超出4张牌就是大
func (baAnswer *BacxAnswerCard) IsBig() bool {

	pnum := baAnswer.GetPlayerNum()
	bnum := baAnswer.GetBankerNum()

	if pnum+bnum >= 5 {
		return true
	}

	return false
}

// 检测是否小
func (baAnswer *BacxAnswerCard) IsSmall() bool {

	pnum := baAnswer.GetPlayerNum()
	bnum := baAnswer.GetBankerNum()

	if pnum+bnum <= 4 {
		return true
	}

	return false
}

type BacxWinner struct {
	Winner int `json:"winner"` //开奖结果  0 闲家， 1 庄家  2 和
	PPair  int `json:ppair`    //闲对子  =1 闲对子
	BPair  int `json:bpair`    //庄对子  =1 庄对子
}

// 开答结构体
type GameOpenAnswerBacx struct {
	GameId string `json:"game_id"` //coolz add 2025.1.24
	DeskNo int    `json:"desk_no"`
	Set    int    `json:"set"`
	Round  int    `json:"round"`

	Winner BacxWinner `json:"winner"`

	BacxAnswerCard `json:"answer_card"`
}

// 历史开答的结构体
type GameBacXAnswersPeriodic struct {
	GameId     string       `json:"game_id"`     //2025.1.6 add
	DeskNo     int          `json:"desk_no"`     //桌号
	Set        int          `json:"set"`         //场次
	Round      int          `json:"round"`       //当前局数
	GameStatus int          `json:"game_status"` //当前游戏状态，参考前面游戏值
	WinnerNum  int          `json:"winner_num"`  //当前开奖数组个数
	Winners    []BacxWinner `json:"winners"`     //已开奖的历史开答路子( 0 闲， 1 庄 ，2 和)
}
