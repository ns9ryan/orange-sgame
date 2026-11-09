package betOdds

import (
	"fmt"
)

// 各游戏的ID
const (
	GType_Bacx = iota
	GType_Bacx_Super6
	GType_Dtx
	GType_Rot
	GType_Sicbo
	GType_Ftan
	GType_Sdie
	GType_Fsc
)

// 压注区域的结构体
type BetOdds struct {
	BetAreaId int //投注区域Id,每个区域唯一
	GameType  int //游戏类型

	Odds   int     //赔率,1賠多少
	Digits int     //賠率小数点位
	OddsF  float64 //赔率的浮点型

	//comment string //说明
}

const (
	BonusWin8or9NRC  = 1 //例牌胜
	BonusWin4Natural = 2 //非例牌4点出
	BonusWin5Natural = 3 //非例牌5点出
	BonusWin6Natural = 4 //非例牌6点出
	BonusWin7Natural = 5 //非例牌7点出
	BonusWin8Natural = 6 //非例牌8点出
	BonusWin9Natural = 7 //非例牌9点出
)

// 用于普通固定赔率的
var betOddsMap map[int]*BetOdds = make(map[int]*BetOdds, 10)

// 专门用于百家乐龙宝赔率
var betOddBonusMap map[int]*BetOdds = make(map[int]*BetOdds, 8)

// 专门用于鱼虾蟹单一点数的赔率，投注每顆骰子1至6中指定的點數 (點數出現一次：1:1,點數出現2二次：1:2，點數出現三次：1:3)
var betOddSinglePointMap map[int]*BetOdds = make(map[int]*BetOdds, 3)

// 初始投注区域的赔率
func init() {

	initDTXOdds()
	initBacxOdds()
	initRotOdds()
	initSicboOdds()
	initFanTanOdds()
	initSdieOdds()
	initFscOdds()

	//初始化区域说明
	initAreadIdCommnet()

	for k, v := range betOddsMap {
		fmt.Printf("betOddsMap[ %d ] %v \n", k, v)
	}
}

// 初始化龙户的区域赔率
func initDTXOdds() {

	//龙区域
	betOddsMap[Bet_Dtx_Dragon] = &BetOdds{Bet_Dtx_Dragon, GType_Dtx, 1, 1, 1}
	//虎区域
	betOddsMap[Bet_Dtx_Tiger] = &BetOdds{Bet_Dtx_Tiger, GType_Dtx, 1, 1, 1}
	//和区域
	betOddsMap[Bet_Dtx_Tie] = &BetOdds{Bet_Dtx_Tie, GType_Dtx, 11, 1, 1}
	//同花顺区域
	betOddsMap[Bet_Dtx_TiePair] = &BetOdds{Bet_Dtx_TiePair, GType_Dtx, 50, 1, 1}

	for k, _ := range betOddsMap {

		if betOddsMap[k].Digits == 0 {
			betOddsMap[k].Digits = 1
		}

		betOddsMap[k].OddsF = float64(betOddsMap[k].Odds / betOddsMap[k].Digits)
	}

}

// 初始化百家乐的区域赔率
func initBacxOdds() {

	betOddsMap[Bet_Bacx_Player] = &BetOdds{Bet_Bacx_Player, GType_Bacx, 1 * 100, 100, 1}
	betOddsMap[Bet_Bacx_Banker] = &BetOdds{Bet_Bacx_Banker, GType_Bacx, 0.95 * 100, 100, 1}
	betOddsMap[Bet_Bacx_Tie] = &BetOdds{Bet_Bacx_Tie, GType_Bacx, 8 * 100, 100, 1}

	betOddsMap[Bet_Bacx_PPair] = &BetOdds{Bet_Bacx_PPair, GType_Bacx, 11 * 100, 100, 1}
	betOddsMap[Bet_Bacx_BPair] = &BetOdds{Bet_Bacx_BPair, GType_Bacx, 11 * 100, 100, 1}

	betOddsMap[Bet_Bacx_BankerNocomm] = &BetOdds{Bet_Bacx_BankerNocomm, GType_Bacx, 1 * 100, 100, 1}
	betOddsMap[Bet_Bacx_SuperSix] = &BetOdds{Bet_Bacx_SuperSix, GType_Bacx, 12 * 100, 100, 1}
	betOddsMap[Bet_Bacx_BankerNocomm] = &BetOdds{Bet_Bacx_BankerNocomm, GType_Bacx, 1 * 100, 100, 1}
	betOddsMap[Bet_Bacx_PerfectPair] = &BetOdds{Bet_Bacx_PerfectPair, GType_Bacx, 25 * 100, 100, 1}
	betOddsMap[Bet_Bacx_AnyPair] = &BetOdds{Bet_Bacx_AnyPair, GType_Bacx, 5 * 100, 100, 1}
	betOddsMap[Bet_Bacx_Big] = &BetOdds{Bet_Bacx_Big, GType_Bacx, 0.5 * 100, 100, 1}
	betOddsMap[Bet_Bacx_Small] = &BetOdds{Bet_Bacx_Small, GType_Bacx, 1.5 * 100, 100, 1}
	betOddsMap[Bet_Bacx_PlayerBonus] = &BetOdds{Bet_Bacx_PlayerBonus, GType_Bacx, 100, 100, 1}
	betOddsMap[Bet_Bacx_BankerBonus] = &BetOdds{Bet_Bacx_BankerBonus, GType_Bacx, 100, 100, 1}

	for k, _ := range betOddsMap {

		if betOddsMap[k].Digits == 0 {
			betOddsMap[k].Digits = 1
		}

		betOddsMap[k].OddsF = float64(betOddsMap[k].Odds / betOddsMap[k].Digits)
	}

	//初始化龙宝的赔率
	betOddBonusMap[BonusWin8or9NRC] = &BetOdds{Bet_Bacx_PlayerBonus, GType_Bacx, 1 * 100, 100, 1}
	betOddBonusMap[BonusWin4Natural] = &BetOdds{Bet_Bacx_PlayerBonus, GType_Bacx, 1 * 100, 100, 1}
	betOddBonusMap[BonusWin5Natural] = &BetOdds{Bet_Bacx_PlayerBonus, GType_Bacx, 2 * 100, 100, 1}
	betOddBonusMap[BonusWin6Natural] = &BetOdds{Bet_Bacx_PlayerBonus, GType_Bacx, 4 * 100, 100, 1}
	betOddBonusMap[BonusWin7Natural] = &BetOdds{Bet_Bacx_PlayerBonus, GType_Bacx, 6 * 100, 100, 1}
	betOddBonusMap[BonusWin8Natural] = &BetOdds{Bet_Bacx_PlayerBonus, GType_Bacx, 10 * 100, 100, 1}
	betOddBonusMap[BonusWin9Natural] = &BetOdds{Bet_Bacx_PlayerBonus, GType_Bacx, 30 * 100, 100, 1}

}

// 初始化轮盘的区域赔率
func initRotOdds() {

	betOddsMap[Bet_Rot_SingleStart] = &BetOdds{Bet_Rot_SingleStart, GType_Rot, 35 * 100, 100, 1}
	betOddsMap[Bet_Rot_DoubleStart] = &BetOdds{Bet_Rot_DoubleStart, GType_Rot, 17 * 100, 100, 1}
	betOddsMap[Bet_Rot_ThreeStart] = &BetOdds{Bet_Rot_ThreeStart, GType_Rot, 11 * 100, 100, 1}
	betOddsMap[Bet_Rot_FourStart] = &BetOdds{Bet_Rot_FourStart, GType_Rot, 8 * 100, 100, 1}
	betOddsMap[Bet_Rot_SixStart] = &BetOdds{Bet_Rot_SixStart, GType_Rot, 5 * 100, 100, 1}
	betOddsMap[Bet_Rot_TwelveStart] = &BetOdds{Bet_Rot_TwelveStart, GType_Rot, 2 * 100, 100, 1}
	betOddsMap[Bet_Rot_EighteenStart] = &BetOdds{Bet_Rot_EighteenStart, GType_Rot, 1 * 100, 100, 1}
	betOddsMap[Bet_Rot_NeighbourTwoStart] = &BetOdds{Bet_Rot_NeighbourTwoStart, GType_Rot, 35 * 100, 100, 1}

	//除了Start的下注区用于结算外，还需要把所有下注区的赔率字段都初始化告诉前端
	for i := Bet_Rot_Single0; i <= Bet_Rot_Single36; i++ {
		betOddsMap[i] = &BetOdds{i, GType_Rot, 35 * 100, 100, 1}
	}

	for i := Bet_Rot_Double1or2; i <= Bet_Rot_Double0or3; i++ {
		betOddsMap[i] = &BetOdds{i, GType_Rot, 17 * 100, 100, 1}
	}

	for i := Bet_Rot_Three1to3; i <= Bet_Rot_Three0or2or3; i++ {
		betOddsMap[i] = &BetOdds{i, GType_Rot, 11 * 100, 100, 1}
	}

	for i := Bet_Rot_Four1or2or4or5; i <= Bet_Rot_Four0or1or2or3; i++ {
		betOddsMap[i] = &BetOdds{i, GType_Rot, 8 * 100, 100, 1}
	}

	for i := Bet_Rot_Six1to6; i <= Bet_Rot_Six31to36; i++ {
		betOddsMap[i] = &BetOdds{i, GType_Rot, 5 * 100, 100, 1}
	}

	for i := Bet_Rot_Twelve1rd12; i <= Bet_Rot_Twelve2to1c; i++ {
		betOddsMap[i] = &BetOdds{i, GType_Rot, 2 * 100, 100, 1}
	}

	for i := Bet_Rot_EighteenRed; i <= Bet_Rot_EighteenDouble; i++ {
		betOddsMap[i] = &BetOdds{i, GType_Rot, 1 * 100, 100, 1}
	}

	for i := Bet_Rot_NeighbourTwo0; i <= Bet_Rot_NeighbourTwo26; i++ {
		betOddsMap[i] = &BetOdds{i, GType_Rot, 35 * 100, 100, 1}
	}

	//默认赔率1陪17
	betOddsMap[Bet_Rot_Tiers] = &BetOdds{Bet_Rot_Tiers, GType_Rot, 17 * 100, 100, 1}

	//注意不开1号默认是1陪17，如果开1号赔率是1陪35，开奖函数内处理
	betOddsMap[Bet_Rot_Orphelins] = &BetOdds{Bet_Rot_Orphelins, GType_Rot, 17 * 100, 100, 1}

	//注意不开0,2,3,25,26等默认赔率是1陪17，否则是1陪11或1陪8，开奖函数内处理
	betOddsMap[Bet_Rot_Voisins] = &BetOdds{Bet_Rot_Voisins, GType_Rot, 17 * 100, 100, 1}

	//注意如果不开26号默认是1陪17，如果开26号是1陪18，开奖函数内处理
	betOddsMap[Bet_Rot_Jeu0] = &BetOdds{Bet_Rot_Jeu0, GType_Rot, 17 * 100, 100, 1}
}

// 初始化筛保区域
func initSicboOdds() {

	betOddsMap[Bet_Sicbo_One_Big] = &BetOdds{Bet_Sicbo_One_Big, GType_Sicbo, 1, 1, 1}
	betOddsMap[Bet_Sicbo_TwoNum_1And2] = &BetOdds{Bet_Sicbo_TwoNum_1And2, GType_Sicbo, 5, 1, 1}
	betOddsMap[Bet_Sicbo_ThreeSame_1All] = &BetOdds{Bet_Sicbo_ThreeSame_1All, GType_Sicbo, 150, 1, 1}
	betOddsMap[Bet_Sicbo_ThreeSame_Any] = &BetOdds{Bet_Sicbo_ThreeSame_Any, GType_Sicbo, 30, 1, 1}

	betOddsMap[Bet_Sicbo_SumPoint_4] = &BetOdds{Bet_Sicbo_SumPoint_4, GType_Sicbo, 50, 1, 1}
	betOddsMap[Bet_Sicbo_SumPoint_5] = &BetOdds{Bet_Sicbo_SumPoint_5, GType_Sicbo, 18, 1, 1}
	betOddsMap[Bet_Sicbo_SumPoint_6] = &BetOdds{Bet_Sicbo_SumPoint_6, GType_Sicbo, 14, 1, 1}
	betOddsMap[Bet_Sicbo_SumPoint_7] = &BetOdds{Bet_Sicbo_SumPoint_7, GType_Sicbo, 12, 1, 1}
	betOddsMap[Bet_Sicbo_SumPoint_8] = &BetOdds{Bet_Sicbo_SumPoint_8, GType_Sicbo, 8, 1, 1}
	betOddsMap[Bet_Sicbo_SumPoint_9] = &BetOdds{Bet_Sicbo_SumPoint_9, GType_Sicbo, 6, 1, 1}
	betOddsMap[Bet_Sicbo_SumPoint_10] = &BetOdds{Bet_Sicbo_SumPoint_10, GType_Sicbo, 6, 1, 1}
	betOddsMap[Bet_Sicbo_SumPoint_11] = &BetOdds{Bet_Sicbo_SumPoint_11, GType_Sicbo, 6, 1, 1}
	betOddsMap[Bet_Sicbo_SumPoint_12] = &BetOdds{Bet_Sicbo_SumPoint_12, GType_Sicbo, 6, 1, 1}
	betOddsMap[Bet_Sicbo_SumPoint_13] = &BetOdds{Bet_Sicbo_SumPoint_13, GType_Sicbo, 8, 1, 1}
	betOddsMap[Bet_Sicbo_SumPoint_14] = &BetOdds{Bet_Sicbo_SumPoint_14, GType_Sicbo, 12, 1, 1}
	betOddsMap[Bet_Sicbo_SumPoint_15] = &BetOdds{Bet_Sicbo_SumPoint_15, GType_Sicbo, 14, 1, 1}
	betOddsMap[Bet_Sicbo_SumPoint_16] = &BetOdds{Bet_Sicbo_SumPoint_16, GType_Sicbo, 18, 1, 1}
	betOddsMap[Bet_Sicbo_SumPoint_17] = &BetOdds{Bet_Sicbo_SumPoint_17, GType_Sicbo, 50, 1, 1}

	betOddsMap[Bet_Sicbo_PairSame_1] = &BetOdds{Bet_Sicbo_PairSame_1, GType_Sicbo, 8, 1, 1}

	betOddsMap[Bet_Sicbo_NumPoint_1] = &BetOdds{Bet_Sicbo_NumPoint_1, GType_Sicbo, 2, 1, 1} //此区域的赔率比较特别，在游戏中实现

}

// 初始化番摊主区域赔率
func initFanTanOdds() {

	//小大单双
	for i := Bet_Ftan_Small_Big_And_Single_Double_Start; i <= Bet_Ftan_Small_Big_And_Single_Double_End; i++ {
		betOddsMap[i] = &BetOdds{i, GType_Ftan, 0.95 * 100, 100, 1}
	}

	//番
	for i := Bet_Ftan_Fan_Start; i <= Bet_Ftan_Fan_End; i++ {
		betOddsMap[i] = &BetOdds{i, GType_Ftan, 2.85 * 100, 100, 1}
	}

	//角
	for i := Bet_Ftan_Jiao_Start; i <= Bet_Ftan_Jiao_End; i++ {
		betOddsMap[i] = &BetOdds{i, GType_Ftan, 0.95 * 100, 100, 1}
	}

	//念
	for i := Bet_Ftan_Nian_Start; i <= Bet_Ftan_Nian_End; i++ {
		betOddsMap[i] = &BetOdds{i, GType_Ftan, 1.9 * 100, 100, 1}
	}

	//三门/三红
	for i := Bet_Ftan_ThreeRed_Start; i <= Bet_Ftan_ThreeRed_End; i++ {
		betOddsMap[i] = &BetOdds{i, GType_Ftan, 0.31 * 100, 100, 1}
	}

	for k, _ := range betOddsMap {

		if betOddsMap[k].Digits == 0 {
			betOddsMap[k].Digits = 1
		}

		betOddsMap[k].OddsF = float64(betOddsMap[k].Odds / betOddsMap[k].Digits)
	}
}

// 初始化色碟主区域赔率
func initSdieOdds() {

	//单双小大(0--4)
	for i := Bet_Sdie_Single; i <= Bet_Sdie_Big; i++ {
		betOddsMap[i] = &BetOdds{i, GType_Sdie, 0.95 * 100, 100, 1}
	}

	//0红4白,4红0白(0--1)
	for i := Bet_Sdie_0Red_4White; i <= Bet_Sdie_4Red_0White; i++ {
		betOddsMap[i] = &BetOdds{i, GType_Sdie, 12 * 100, 100, 1}
	}

	//1红3白,3红1白(0--1)
	for i := Bet_Sdie_1Red_3White; i <= Bet_Sdie_3Red_1White; i++ {
		betOddsMap[i] = &BetOdds{i, GType_Sdie, 2.6 * 100, 100, 1}
	}

	betOddsMap[Bet_Sdie_2Red_2White] = &BetOdds{Bet_Sdie_2Red_2White, GType_Sdie, 1.5 * 100, 100, 1}       //2白2红(0)
	betOddsMap[Bet_Sdie_4White_Or_4Red] = &BetOdds{Bet_Sdie_4White_Or_4Red, GType_Sdie, 6.5 * 100, 100, 1} //4白或4红(0)

	for k, _ := range betOddsMap {

		if betOddsMap[k].Digits == 0 {
			betOddsMap[k].Digits = 1
		}

		betOddsMap[k].OddsF = float64(betOddsMap[k].Odds / betOddsMap[k].Digits)
	}
}

// 得到投注区域对应的赔率
func GetBetOddsF(areaId int) float64 {

	if (areaId >= Bet_Area_Start) && (areaId <= Bet_Area_End) {

		v, ok := betOddsMap[areaId]

		if ok == false {
			return 0
		}

		return v.OddsF
	}

	return 0
}

// 得到投注区域对应的赔率
func GetBetOdds(areaId int) (odds, digits int) {

	if (areaId >= Bet_Area_Start) && (areaId <= Bet_Area_End) {

		v, ok := betOddsMap[areaId]

		if ok == false {
			odds = 0
			digits = 1

			return odds, digits
		}

		return v.Odds, v.Digits
	}

	return 0, 1
}

// 百家乐龙宝投注区域的赔率
func GetBonusOdds(bonusType int) (odds, digits int) {

	//检测是否在这个区域内
	if !(bonusType >= BonusWin8or9NRC && bonusType <= BonusWin9Natural) {

		return 0, 1
	}

	odds = betOddBonusMap[bonusType].Odds
	digits = betOddBonusMap[bonusType].Digits

	return odds, digits
}

// 鱼虾蟹押单一点数的赔率
func initFscOdds() {
	//初始化单一点数下注区的Digits，用来返回给前端（字段Odds，OddsF无用到，此押注区比较特殊，一个区域的赔率，根据开奖结果才能确定的）
	betOddsMap[Bet_Fsc_Single_Point1] = &BetOdds{Bet_Fsc_Single_Point1, GType_Fsc, 1, 1, 1}
	betOddsMap[Bet_Fsc_Single_Point2] = &BetOdds{Bet_Fsc_Single_Point2, GType_Fsc, 1, 1, 1}
	betOddsMap[Bet_Fsc_Single_Point3] = &BetOdds{Bet_Fsc_Single_Point3, GType_Fsc, 1, 1, 1}
	betOddsMap[Bet_Fsc_Single_Point4] = &BetOdds{Bet_Fsc_Single_Point4, GType_Fsc, 1, 1, 1}
	betOddsMap[Bet_Fsc_Single_Point5] = &BetOdds{Bet_Fsc_Single_Point5, GType_Fsc, 1, 1, 1}
	betOddsMap[Bet_Fsc_Single_Point6] = &BetOdds{Bet_Fsc_Single_Point6, GType_Fsc, 1, 1, 1}

	// 1骰，2骰，全骰对应的赔率，计算的时候用到
	betOddSinglePointMap[1] = &BetOdds{0, GType_Fsc, 2 * 1, 1, 1}
	betOddSinglePointMap[2] = &BetOdds{0, GType_Fsc, 3 * 1, 1, 1}
	betOddSinglePointMap[3] = &BetOdds{0, GType_Fsc, 30 * 1, 1, 1}
}

// 获取鱼虾蟹单一点数的赔率
func GetFscSinglePointOdds(num int) (odds, digits int) {

	if num == 0 {
		return 0, 1
	}

	odds = betOddSinglePointMap[num].Odds
	digits = betOddSinglePointMap[num].Digits

	return odds, digits
}
