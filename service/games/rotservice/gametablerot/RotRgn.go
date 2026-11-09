package gameTableRot

import (
	"fmt"
	betOdds "videogame/common/betodds"
)

//
//=============================================================================================
// 轮盘投注区域
// 主区域代号： 单数：0，双数：1，三数：2，四数：3，六数：4，12数：5，红黑大小单双：6等12个主区域
// 亚区域：投注的数组索引
//=============================================================================================

const (
	SingleIdNum   = 37
	DoubeIdNum    = 60
	ThreeIdNum    = 14
	FourIdNum     = 23
	SixIdNum      = 11
	TwelveIdNum   = 6
	EighteenIdNum = 6
	NeighbourTwoIdNum = 37
	TiersIdNum = 1
	OrphelinsIdNum = 1
	VoisinsIdNum = 1
	Jeu0IdNum = 1

	MaxIdNum      = SingleIdNum + DoubeIdNum + ThreeIdNum + FourIdNum + SixIdNum + TwelveIdNum + EighteenIdNum + NeighbourTwoIdNum + TiersIdNum + OrphelinsIdNum + VoisinsIdNum + Jeu0IdNum


	//定义当在轮盘圈上投注时，下注筹码的倍数要求
	NeighbourTwoMulti = 5
	TiersMulti = 6
	OrphelinsMulti = 5
	VoisinsMulti = 9
	Jeu0Multi = 4

)

var DoubleRots [DoubeIdNum][2]int
var ThreeRots [ThreeIdNum][3]int
var FourRots [FourIdNum][4]int
var SixRots [SixIdNum][6]int
var TwelveRots [TwelveIdNum][12]int
var EighteenRots [EighteenIdNum][18]int
var NeighbourTwoRots [NeighbourTwoIdNum][5]int

// 用于快速检查中奖号码有没有在轮盘下角注/轮盘孤注/零旁注上角/轮上零旁区域里
var TiersRange map[int]struct{}
var OrphelinsRange map[int]struct{}
var VoisinsRotsRange map[int]struct{}
var Jeu0Range map[int]struct{}

// 用于读取有中奖的区域id列表用于前端高亮显示
var WinnerIdsMap map[int][]uint32

func init() {

	//单数(1区) -- 36       1赔35         ( rgnMain: 0, rgnsub 0--- 35 )
	singleRots := [37]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36}

	//双数(2区)  1赔17   ( rgnMain:1   rgnsub: 0 -- 59   压双数 )
	doubleRots := [60][2]int{{1, 2}, {2, 3}, {4, 5}, {5, 6}, {7, 8}, {8, 9}, {10, 11}, {11, 12}, {13, 14}, {14, 15}, {16, 17}, {17, 18},
		{19, 20}, {20, 21}, {22, 23}, {23, 24}, {25, 26}, {26, 27}, {28, 29}, {29, 30}, {31, 32}, {32, 33}, {34, 35}, {35, 36},
		{1, 4}, {4, 7}, {7, 10}, {10, 13}, {13, 16}, {16, 19}, {19, 22}, {22, 25}, {25, 28}, {28, 31}, {31, 34},
		{2, 5}, {5, 8}, {8, 11}, {11, 14}, {14, 17}, {17, 20}, {20, 23}, {23, 26}, {26, 29}, {29, 32}, {32, 35},
		{3, 6}, {6, 9}, {9, 12}, {12, 15}, {15, 18}, {18, 21}, {21, 24}, {24, 27}, {27, 30}, {30, 33}, {33, 36},
		{0, 1}, {0, 2}, {0, 3}}

	//三数(3区)  1赔11   ( rgnMain:2   rgnsub: 0 -- 13   压三数 )
	threeRots := [14][3]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}, {10, 11, 12}, {13, 14, 15}, {16, 17, 18},
		{19, 20, 21}, {22, 23, 24}, {25, 26, 27}, {28, 29, 30}, {31, 32, 33}, {34, 35, 36},
		{0, 1, 2}, {0, 2, 3}}

	//四数(4区)  1赔8 ( rgnMain:3   rgnsub: 0 -- 22   压四个数的 )
	fourRots := [23][4]int{{1, 2, 4, 5}, {2, 3, 5, 6}, {4, 5, 7, 8}, {5, 6, 8, 9}, {7, 8, 10, 11}, {8, 9, 11, 12}, {10, 11, 13, 14}, {11, 12, 14, 15},
		{13, 14, 16, 17}, {14, 15, 17, 18}, {16, 17, 19, 20}, {17, 18, 20, 21}, {19, 20, 22, 23}, {20, 21, 23, 24}, {22, 23, 25, 26},
		{23, 24, 26, 27}, {25, 26, 28, 29}, {26, 27, 29, 30}, {28, 29, 31, 32}, {29, 30, 32, 33}, {31, 32, 34, 35}, {32, 33, 35, 36},
		{0, 1, 2, 3}}

	//六数(5区)	1赔5 ( rgnMain:4   rgnsub: 0 -- 10   //压6个数 )
	sixRots := [11][6]int{{1, 2, 3, 4, 5, 6}, {4, 5, 6, 7, 8, 9}, {7, 8, 9, 10, 11, 12}, {10, 11, 12, 13, 14, 15}, {13, 14, 15, 16, 17, 18}, {16, 17, 18, 19, 20, 21},
		{19, 20, 21, 22, 23, 24}, {22, 23, 24, 25, 26, 27}, {25, 26, 27, 28, 29, 30}, {28, 29, 30, 31, 32, 33}, {31, 32, 33, 34, 35, 36}}

	//12数(6区)	1赔2 ( rgnMain:5   rgnsub: 0 -- 5  压12个数  )
	twelveRots := [6][12]int{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, {13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24}, {25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36},
		{1, 4, 7, 10, 13, 16, 19, 22, 25, 28, 31, 34}, {2, 5, 8, 11, 14, 17, 20, 23, 26, 29, 32, 35}, {3, 6, 9, 12, 15, 18, 21, 24, 27, 30, 33, 36}}

	//红黑小大单双(7区)  1赔1 ( rgnMain:6   rgnsub: 0 -- 5 压18个数  )
	eighteenRots := [6][18]int{{1, 3, 5, 7, 9, 12, 14, 16, 18, 19, 21, 23, 25, 27, 30, 32, 34, 36}, {2, 4, 6, 8, 10, 11, 13, 15, 17, 20, 22, 24, 26, 28, 29, 31, 33, 35},
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}, {19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36},
		{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31, 33, 35}, {2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36}}

	//邻注(8区) 两邻居共5个数，需5倍筹码，1赔35 ( rgnMain:7   rgnsub: 0 -- 36//压5个数 )，以下排列从0顺时针
	neighbourTwoRots := [37][5]int{{15, 32, 0, 26, 3}, {19, 15, 32, 0, 26}, {4, 19, 15, 32, 0}, {21, 4, 19, 15, 32}, {2, 21, 4, 19, 15}, {25, 2, 21, 4, 19}, {17, 25, 2, 21, 4}, {34, 17, 25, 2, 21}, {6, 34, 17, 25, 2}, {27, 6, 34, 17, 25}, {13, 27, 6, 34, 17}, {36, 13, 27, 6, 34}, {11, 36, 13, 27, 6}, {30, 11, 36, 13, 27}, {8, 30, 11, 36, 13}, {23, 8, 30, 11, 36}, {10, 23, 8, 30, 11}, {5, 10, 23, 8, 30}, {24, 5, 10, 23, 8}, {16, 24, 5, 10, 23}, {33, 16, 24, 5, 10}, {1, 33, 16, 24, 5}, {20, 1, 33, 16, 24}, {14, 20, 1, 33, 16}, {31, 14, 20, 1, 33}, {9, 31, 14, 20, 1}, {22, 9, 31, 14, 20}, {18, 22, 9, 31, 14}, {29, 18, 22, 9, 31}, {7, 29, 18, 22, 9}, {28, 7, 29, 18, 22}, {12, 28, 7, 29, 18}, {35, 12, 28, 7, 29}, {3, 35, 12, 28, 7}, {26, 3, 35, 12, 28}, {0, 26, 3, 35, 12}, {32, 0, 26, 3, 35}}


	_ = singleRots
	DoubleRots = doubleRots
	ThreeRots = threeRots
	FourRots = fourRots
	SixRots = sixRots
	TwelveRots = twelveRots
	EighteenRots = eighteenRots
	NeighbourTwoRots = neighbourTwoRots

	// 5  24 16 33 1  20 14 31 9  22 18 29 7  28 12 35
	// 10                                            3
	// 23  Tiers    \Orphelins    /Voisins     /Jeu0 26
	// 8                                             0
	// 30 11 36 13 27 6  34 17 25 2  21 4  19 15 32

	//轮盘下角注(9区)，包含轮盘圈上的12个数字，需6倍筹码，分解成6个双数注，见betodds说明
	TiersRange = map[int]struct{}{30: {}, 11: {}, 36: {}, 13: {}, 27: {}, 33: {}, 16: {}, 24: {}, 5: {}, 10: {}, 23: {}, 8: {}}

	//轮盘孤注(10区)，包含轮盘圈上的8个数字，需5倍筹码，分解成1个单注和4个双注，见betodds说明
	OrphelinsRange = map[int]struct{}{6: {}, 34: {}, 17: {}, 9: {}, 31: {}, 14: {}, 20: {}, 1: {}}

	//零旁注上角(11区)，包含轮盘圈上的17个数字，需9倍筹码，分解成好几个注，见betodds说明
	VoisinsRotsRange = map[int]struct{}{25: {}, 2: {}, 21: {}, 4: {}, 19: {}, 15: {}, 32: {}, 0: {}, 26: {}, 3: {}, 35: {}, 12: {}, 28: {}, 7: {}, 29: {}, 18: {}, 22: {}}

	//轮上零旁(12区)，包含轮盘圈上的6个数字，需4倍筹码，分解成1个单注和3个双注，，见betodds说明
	Jeu0Range = map[int]struct{}{15: {}, 32: {}, 0: {}, 26: {}, 3: {}, 35: {}, 12: {}}

	//用于读取有中奖的区域id列表用于前端高亮显示，一共0-36个固定值映射
	WinnerIdsMap = make(map[int][]uint32)
	WinnerIdsMap[0] = []uint32{
		betOdds.Bet_Rot_Single0,
		betOdds.Bet_Rot_NeighbourTwo0,
		betOdds.Bet_Rot_Voisins,
		betOdds.Bet_Rot_Jeu0,
	}

	WinnerIdsMap[1] = []uint32{
		betOdds.Bet_Rot_Single1,
		betOdds.Bet_Rot_Twelve1rd12,
		betOdds.Bet_Rot_Twelve2to1a,
		betOdds.Bet_Rot_EighteenRed,
		betOdds.Bet_Rot_EighteenSmall,
		betOdds.Bet_Rot_EighteenSingle,
		betOdds.Bet_Rot_NeighbourTwo1,
		betOdds.Bet_Rot_Orphelins,
	}

	WinnerIdsMap[2] = []uint32{
		betOdds.Bet_Rot_Single2,
		betOdds.Bet_Rot_Twelve1rd12,
		betOdds.Bet_Rot_Twelve2to1b,
		betOdds.Bet_Rot_EighteenBlack,
		betOdds.Bet_Rot_EighteenSmall,
		betOdds.Bet_Rot_EighteenDouble,
		betOdds.Bet_Rot_NeighbourTwo2,
		betOdds.Bet_Rot_Voisins,
	}

	WinnerIdsMap[3] = []uint32{
		betOdds.Bet_Rot_Single3,
		betOdds.Bet_Rot_Twelve1rd12,
		betOdds.Bet_Rot_Twelve2to1c,
		betOdds.Bet_Rot_EighteenRed,
		betOdds.Bet_Rot_EighteenSmall,
		betOdds.Bet_Rot_EighteenSingle,
		betOdds.Bet_Rot_NeighbourTwo3,
		betOdds.Bet_Rot_Voisins,
		betOdds.Bet_Rot_Jeu0,
	}

	WinnerIdsMap[4] = []uint32{
		betOdds.Bet_Rot_Single4,
		betOdds.Bet_Rot_Twelve1rd12,
		betOdds.Bet_Rot_Twelve2to1a,
		betOdds.Bet_Rot_EighteenBlack,
		betOdds.Bet_Rot_EighteenSmall,
		betOdds.Bet_Rot_EighteenDouble,
		betOdds.Bet_Rot_NeighbourTwo4,
		betOdds.Bet_Rot_Voisins,
	}

	WinnerIdsMap[5] = []uint32{
		betOdds.Bet_Rot_Single5,
		betOdds.Bet_Rot_Twelve1rd12,
		betOdds.Bet_Rot_Twelve2to1b,
		betOdds.Bet_Rot_EighteenRed,
		betOdds.Bet_Rot_EighteenSmall,
		betOdds.Bet_Rot_EighteenSingle,
		betOdds.Bet_Rot_NeighbourTwo5,
		betOdds.Bet_Rot_Tiers,
	}

	WinnerIdsMap[6] = []uint32{
		betOdds.Bet_Rot_Single6,
		betOdds.Bet_Rot_Twelve1rd12,
		betOdds.Bet_Rot_Twelve2to1c,
		betOdds.Bet_Rot_EighteenBlack,
		betOdds.Bet_Rot_EighteenSmall,
		betOdds.Bet_Rot_EighteenDouble,
		betOdds.Bet_Rot_NeighbourTwo6,
		betOdds.Bet_Rot_Orphelins,
	}

	WinnerIdsMap[7] = []uint32{
		betOdds.Bet_Rot_Single7,
		betOdds.Bet_Rot_Twelve1rd12,
		betOdds.Bet_Rot_Twelve2to1a,
		betOdds.Bet_Rot_EighteenRed,
		betOdds.Bet_Rot_EighteenSmall,
		betOdds.Bet_Rot_EighteenSingle,
		betOdds.Bet_Rot_NeighbourTwo7,
		betOdds.Bet_Rot_Voisins,
	}

	WinnerIdsMap[8] = []uint32{
		betOdds.Bet_Rot_Single8,
		betOdds.Bet_Rot_Twelve1rd12,
		betOdds.Bet_Rot_Twelve2to1b,
		betOdds.Bet_Rot_EighteenBlack,
		betOdds.Bet_Rot_EighteenSmall,
		betOdds.Bet_Rot_EighteenDouble,
		betOdds.Bet_Rot_NeighbourTwo8,
		betOdds.Bet_Rot_Tiers,
	}

	WinnerIdsMap[9] = []uint32{
		betOdds.Bet_Rot_Single9,
		betOdds.Bet_Rot_Twelve1rd12,
		betOdds.Bet_Rot_Twelve2to1c,
		betOdds.Bet_Rot_EighteenRed,
		betOdds.Bet_Rot_EighteenSmall,
		betOdds.Bet_Rot_EighteenSingle,
		betOdds.Bet_Rot_NeighbourTwo9,
		betOdds.Bet_Rot_Orphelins,
	}

	WinnerIdsMap[10] = []uint32{
		betOdds.Bet_Rot_Single10,
		betOdds.Bet_Rot_Twelve1rd12,
		betOdds.Bet_Rot_Twelve2to1a,
		betOdds.Bet_Rot_EighteenBlack,
		betOdds.Bet_Rot_EighteenSmall,
		betOdds.Bet_Rot_EighteenSingle,
		betOdds.Bet_Rot_NeighbourTwo10,
		betOdds.Bet_Rot_Tiers,
	}

	WinnerIdsMap[11] = []uint32{
		betOdds.Bet_Rot_Single11,
		betOdds.Bet_Rot_Twelve1rd12,
		betOdds.Bet_Rot_Twelve2to1b,
		betOdds.Bet_Rot_EighteenBlack,
		betOdds.Bet_Rot_EighteenSmall,
		betOdds.Bet_Rot_EighteenSingle,
		betOdds.Bet_Rot_NeighbourTwo11,
		betOdds.Bet_Rot_Tiers,
	}

	WinnerIdsMap[12] = []uint32{
		betOdds.Bet_Rot_Single12,
		betOdds.Bet_Rot_Twelve1rd12,
		betOdds.Bet_Rot_Twelve2to1c,
		betOdds.Bet_Rot_EighteenRed,
		betOdds.Bet_Rot_EighteenSmall,
		betOdds.Bet_Rot_EighteenDouble,
		betOdds.Bet_Rot_NeighbourTwo12,
		betOdds.Bet_Rot_Voisins,
		betOdds.Bet_Rot_Jeu0,
	}

	WinnerIdsMap[13] = []uint32{
		betOdds.Bet_Rot_Single13,
		betOdds.Bet_Rot_Twelve2nd12,
		betOdds.Bet_Rot_Twelve2to1a,
		betOdds.Bet_Rot_EighteenBlack,
		betOdds.Bet_Rot_EighteenSmall,
		betOdds.Bet_Rot_EighteenSingle,
		betOdds.Bet_Rot_NeighbourTwo13,
		betOdds.Bet_Rot_Tiers,
	}

	WinnerIdsMap[14] = []uint32{
		betOdds.Bet_Rot_Single14,
		betOdds.Bet_Rot_Twelve2nd12,
		betOdds.Bet_Rot_Twelve2to1b,
		betOdds.Bet_Rot_EighteenRed,
		betOdds.Bet_Rot_EighteenSmall,
		betOdds.Bet_Rot_EighteenDouble,
		betOdds.Bet_Rot_NeighbourTwo14,
		betOdds.Bet_Rot_Orphelins,
	}

	WinnerIdsMap[15] = []uint32{
		betOdds.Bet_Rot_Single15,
		betOdds.Bet_Rot_Twelve2nd12,
		betOdds.Bet_Rot_Twelve2to1c,
		betOdds.Bet_Rot_EighteenBlack,
		betOdds.Bet_Rot_EighteenSmall,
		betOdds.Bet_Rot_EighteenSingle,
		betOdds.Bet_Rot_NeighbourTwo15,
		betOdds.Bet_Rot_Voisins,
		betOdds.Bet_Rot_Jeu0,
	}

	WinnerIdsMap[16] = []uint32{
		betOdds.Bet_Rot_Single16,
		betOdds.Bet_Rot_Twelve2nd12,
		betOdds.Bet_Rot_Twelve2to1a,
		betOdds.Bet_Rot_EighteenRed,
		betOdds.Bet_Rot_EighteenSmall,
		betOdds.Bet_Rot_EighteenDouble,
		betOdds.Bet_Rot_NeighbourTwo16,
		betOdds.Bet_Rot_Tiers,
	}

	WinnerIdsMap[17] = []uint32{
		betOdds.Bet_Rot_Single17,
		betOdds.Bet_Rot_Twelve2nd12,
		betOdds.Bet_Rot_Twelve2to1b,
		betOdds.Bet_Rot_EighteenBlack,
		betOdds.Bet_Rot_EighteenSmall,
		betOdds.Bet_Rot_EighteenSingle,
		betOdds.Bet_Rot_NeighbourTwo17,
		betOdds.Bet_Rot_Orphelins,
	}

	WinnerIdsMap[18] = []uint32{
		betOdds.Bet_Rot_Single18,
		betOdds.Bet_Rot_Twelve2nd12,
		betOdds.Bet_Rot_Twelve2to1c,
		betOdds.Bet_Rot_EighteenRed,
		betOdds.Bet_Rot_EighteenSmall,
		betOdds.Bet_Rot_EighteenDouble,
		betOdds.Bet_Rot_NeighbourTwo18,
		betOdds.Bet_Rot_Voisins,
	}

	WinnerIdsMap[19] = []uint32{
		betOdds.Bet_Rot_Single19,
		betOdds.Bet_Rot_Twelve2nd12,
		betOdds.Bet_Rot_Twelve2to1a,
		betOdds.Bet_Rot_EighteenRed,
		betOdds.Bet_Rot_EighteenBig,
		betOdds.Bet_Rot_EighteenSingle,
		betOdds.Bet_Rot_NeighbourTwo19,
		betOdds.Bet_Rot_Voisins,
	}

	WinnerIdsMap[20] = []uint32{
		betOdds.Bet_Rot_Single20,
		betOdds.Bet_Rot_Twelve2nd12,
		betOdds.Bet_Rot_Twelve2to1b,
		betOdds.Bet_Rot_EighteenBlack,
		betOdds.Bet_Rot_EighteenBig,
		betOdds.Bet_Rot_EighteenDouble,
		betOdds.Bet_Rot_NeighbourTwo20,
		betOdds.Bet_Rot_Orphelins,
	}

	WinnerIdsMap[21] = []uint32{
		betOdds.Bet_Rot_Single21,
		betOdds.Bet_Rot_Twelve2nd12,
		betOdds.Bet_Rot_Twelve2to1c,
		betOdds.Bet_Rot_EighteenRed,
		betOdds.Bet_Rot_EighteenBig,
		betOdds.Bet_Rot_EighteenSingle,
		betOdds.Bet_Rot_NeighbourTwo21,
		betOdds.Bet_Rot_Voisins,
	}

	WinnerIdsMap[22] = []uint32{
		betOdds.Bet_Rot_Single22,
		betOdds.Bet_Rot_Twelve2nd12,
		betOdds.Bet_Rot_Twelve2to1a,
		betOdds.Bet_Rot_EighteenBlack,
		betOdds.Bet_Rot_EighteenBig,
		betOdds.Bet_Rot_EighteenDouble,
		betOdds.Bet_Rot_NeighbourTwo22,
		betOdds.Bet_Rot_Voisins,
	}

	WinnerIdsMap[23] = []uint32{
		betOdds.Bet_Rot_Single23,
		betOdds.Bet_Rot_Twelve2nd12,
		betOdds.Bet_Rot_Twelve2to1b,
		betOdds.Bet_Rot_EighteenRed,
		betOdds.Bet_Rot_EighteenBig,
		betOdds.Bet_Rot_EighteenSingle,
		betOdds.Bet_Rot_NeighbourTwo23,
		betOdds.Bet_Rot_Tiers,
	}

	WinnerIdsMap[24] = []uint32{
		betOdds.Bet_Rot_Single24,
		betOdds.Bet_Rot_Twelve2nd12,
		betOdds.Bet_Rot_Twelve2to1c,
		betOdds.Bet_Rot_EighteenBlack,
		betOdds.Bet_Rot_EighteenBig,
		betOdds.Bet_Rot_EighteenDouble,
		betOdds.Bet_Rot_NeighbourTwo24,
		betOdds.Bet_Rot_Tiers,
	}

	WinnerIdsMap[25] = []uint32{
		betOdds.Bet_Rot_Single25,
		betOdds.Bet_Rot_Twelve3rd12,
		betOdds.Bet_Rot_Twelve2to1a,
		betOdds.Bet_Rot_EighteenRed,
		betOdds.Bet_Rot_EighteenBig,
		betOdds.Bet_Rot_EighteenSingle,
		betOdds.Bet_Rot_NeighbourTwo25,
		betOdds.Bet_Rot_Voisins,
	}

	WinnerIdsMap[26] = []uint32{
		betOdds.Bet_Rot_Single26,
		betOdds.Bet_Rot_Twelve3rd12,
		betOdds.Bet_Rot_Twelve2to1b,
		betOdds.Bet_Rot_EighteenBlack,
		betOdds.Bet_Rot_EighteenBig,
		betOdds.Bet_Rot_EighteenDouble,
		betOdds.Bet_Rot_NeighbourTwo26,
		betOdds.Bet_Rot_Voisins,
		betOdds.Bet_Rot_Jeu0,
	}

	WinnerIdsMap[27] = []uint32{
		betOdds.Bet_Rot_Single27,
		betOdds.Bet_Rot_Twelve3rd12,
		betOdds.Bet_Rot_Twelve2to1c,
		betOdds.Bet_Rot_EighteenRed,
		betOdds.Bet_Rot_EighteenBig,
		betOdds.Bet_Rot_EighteenSingle,
		betOdds.Bet_Rot_NeighbourTwo27,
		betOdds.Bet_Rot_Tiers,
	}

	WinnerIdsMap[28] = []uint32{
		betOdds.Bet_Rot_Single28,
		betOdds.Bet_Rot_Twelve3rd12,
		betOdds.Bet_Rot_Twelve2to1a,
		betOdds.Bet_Rot_EighteenBlack,
		betOdds.Bet_Rot_EighteenBig,
		betOdds.Bet_Rot_EighteenDouble,
		betOdds.Bet_Rot_NeighbourTwo28,
		betOdds.Bet_Rot_Voisins,
	}

	WinnerIdsMap[29] = []uint32{
		betOdds.Bet_Rot_Single29,
		betOdds.Bet_Rot_Twelve3rd12,
		betOdds.Bet_Rot_Twelve2to1b,
		betOdds.Bet_Rot_EighteenBlack,
		betOdds.Bet_Rot_EighteenBig,
		betOdds.Bet_Rot_EighteenSingle,
		betOdds.Bet_Rot_NeighbourTwo29,
		betOdds.Bet_Rot_Voisins,
	}

	WinnerIdsMap[30] = []uint32{
		betOdds.Bet_Rot_Single30,
		betOdds.Bet_Rot_Twelve3rd12,
		betOdds.Bet_Rot_Twelve2to1c,
		betOdds.Bet_Rot_EighteenRed,
		betOdds.Bet_Rot_EighteenBig,
		betOdds.Bet_Rot_EighteenDouble,
		betOdds.Bet_Rot_NeighbourTwo30,
		betOdds.Bet_Rot_Tiers,
	}

	WinnerIdsMap[31] = []uint32{
		betOdds.Bet_Rot_Single31,
		betOdds.Bet_Rot_Twelve3rd12,
		betOdds.Bet_Rot_Twelve2to1a,
		betOdds.Bet_Rot_EighteenBlack,
		betOdds.Bet_Rot_EighteenBig,
		betOdds.Bet_Rot_EighteenSingle,
		betOdds.Bet_Rot_NeighbourTwo31,
		betOdds.Bet_Rot_Orphelins,
	}

	WinnerIdsMap[32] = []uint32{
		betOdds.Bet_Rot_Single32,
		betOdds.Bet_Rot_Twelve3rd12,
		betOdds.Bet_Rot_Twelve2to1b,
		betOdds.Bet_Rot_EighteenRed,
		betOdds.Bet_Rot_EighteenBig,
		betOdds.Bet_Rot_EighteenDouble,
		betOdds.Bet_Rot_NeighbourTwo32,
		betOdds.Bet_Rot_Voisins,
		betOdds.Bet_Rot_Jeu0,
	}

	WinnerIdsMap[33] = []uint32{
		betOdds.Bet_Rot_Single33,
		betOdds.Bet_Rot_Twelve3rd12,
		betOdds.Bet_Rot_Twelve2to1c,
		betOdds.Bet_Rot_EighteenBlack,
		betOdds.Bet_Rot_EighteenBig,
		betOdds.Bet_Rot_EighteenSingle,
		betOdds.Bet_Rot_NeighbourTwo33,
		betOdds.Bet_Rot_Tiers,
	}

	WinnerIdsMap[34] = []uint32{
		betOdds.Bet_Rot_Single34,
		betOdds.Bet_Rot_Twelve3rd12,
		betOdds.Bet_Rot_Twelve2to1a,
		betOdds.Bet_Rot_EighteenRed,
		betOdds.Bet_Rot_EighteenBig,
		betOdds.Bet_Rot_EighteenDouble,
		betOdds.Bet_Rot_NeighbourTwo34,
		betOdds.Bet_Rot_Orphelins,
	}

	WinnerIdsMap[35] = []uint32{
		betOdds.Bet_Rot_Single35,
		betOdds.Bet_Rot_Twelve3rd12,
		betOdds.Bet_Rot_Twelve2to1b,
		betOdds.Bet_Rot_EighteenBlack,
		betOdds.Bet_Rot_EighteenBig,
		betOdds.Bet_Rot_EighteenSingle,
		betOdds.Bet_Rot_NeighbourTwo35,
		betOdds.Bet_Rot_Voisins,
		betOdds.Bet_Rot_Jeu0,
	}

	WinnerIdsMap[36] = []uint32{
		betOdds.Bet_Rot_Single36,
		betOdds.Bet_Rot_Twelve3rd12,
		betOdds.Bet_Rot_Twelve2to1c,
		betOdds.Bet_Rot_EighteenRed,
		betOdds.Bet_Rot_EighteenBig,
		betOdds.Bet_Rot_EighteenDouble,
		betOdds.Bet_Rot_NeighbourTwo36,
		betOdds.Bet_Rot_Tiers,
	}

	// fmt.Printf("***%v \n", DoubleRots)
	// fmt.Printf("***%v \n", ThreeRots)
	// fmt.Printf("***%v \n", FourRots)
	// fmt.Printf("***%v \n", SixRots)
	// fmt.Printf("***%v \n", TwelveRots)
	// fmt.Printf("***%v \n", EighteenRots)
	// fmt.Printf("***%v \n", NeighbourTwoRots)

	// fmt.Printf("***%v \n", TiersRange)
	// fmt.Printf("***%v \n", OrphelinsRange)
	// fmt.Printf("***%v \n", VoisinsRotsRange)
	// fmt.Printf("***%v \n", Jeu0Range)

	fmt.Printf("***%v \n", WinnerIdsMap)
}

// 投注区的二维结构体
type RotRgn struct {
	RgnMain int
	RgnSub  int
}

func (rot *RotRgn) IsSame() {

}

// 轮盘赌红色号码
var redNumbers = map[int]bool{
	1: true, 3: true, 5: true, 7: true, 9: true, 12: true, 14: true, 16: true, 18: true, 19: true, 21: true, 23: true, 25: true, 27: true, 30: true, 32: true, 34: true, 36: true,
}

// 判断是否为红色
func IsRed(number int) int {
	if value, ok := redNumbers[number]; ok && value {
		return 1
	}
	return 0
}

// 判断是否为大（19-36为大）
func IsBig(number int) int {
	if number >= 19 && number <= 36 {
		return 1
	}
	return 0
}

// 判断是否为单数
func IsSingle(number int) int {
	if number%2 != 0 {
		return 1
	}
	return 0
}
