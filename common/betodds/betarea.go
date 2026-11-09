package betOdds

// 全部投注区域的ID
const (
	Bet_Area_Start = 1000

	Bet_Bacx_Start = Bet_Area_Start + 1

	Bet_Bacx_Player = Bet_Bacx_Start + 0
	Bet_Bacx_Banker = Bet_Bacx_Start + 1
	Bet_Bacx_Tie    = Bet_Bacx_Start + 2
	Bet_Bacx_PPair  = Bet_Bacx_Start + 3
	Bet_Bacx_BPair  = Bet_Bacx_Start + 4

	Bet_Bacx_BankerNocomm = Bet_Bacx_Start + 5
	Bet_Bacx_SuperSix     = Bet_Bacx_Start + 6
	Bet_Bacx_AnyPair      = Bet_Bacx_Start + 7
	Bet_Bacx_PerfectPair  = Bet_Bacx_Start + 8
	Bet_Bacx_Big          = Bet_Bacx_Start + 9
	Bet_Bacx_Small        = Bet_Bacx_Start + 10
	Bet_Bacx_PlayerBonus  = Bet_Bacx_Start + 11
	Bet_Bacx_BankerBonus  = Bet_Bacx_Start + 12

	/**************************************************************************/

	Bet_Dtx_Start = Bet_Bacx_BankerBonus + 1 //   Bet_Bacx_BankerBonus + 1

	Bet_Dtx_Dragon  = Bet_Dtx_Start + 0 //龙区域
	Bet_Dtx_Tiger   = Bet_Dtx_Start + 1 //闲区域
	Bet_Dtx_Tie     = Bet_Dtx_Start + 2 //和区域
	Bet_Dtx_TiePair = Bet_Dtx_Start + 3 //同花和局

	/**************************************************************************/

	Bet_Rot_Start = 1100 //12个不同赔率大区，198个下注点

	//单数(1区) -- 36       1赔35         ( rgnMain: 0, rgnsub 0--- 35 )
	Bet_Rot_SingleStart = Bet_Rot_Start

	Bet_Rot_Single0  = Bet_Rot_SingleStart + 0
	Bet_Rot_Single1  = Bet_Rot_SingleStart + 1
	Bet_Rot_Single2  = Bet_Rot_SingleStart + 2
	Bet_Rot_Single3  = Bet_Rot_SingleStart + 3
	Bet_Rot_Single4  = Bet_Rot_SingleStart + 4
	Bet_Rot_Single5  = Bet_Rot_SingleStart + 5
	Bet_Rot_Single6  = Bet_Rot_SingleStart + 6
	Bet_Rot_Single7  = Bet_Rot_SingleStart + 7
	Bet_Rot_Single8  = Bet_Rot_SingleStart + 8
	Bet_Rot_Single9  = Bet_Rot_SingleStart + 9
	Bet_Rot_Single10 = Bet_Rot_SingleStart + 10
	Bet_Rot_Single11 = Bet_Rot_SingleStart + 11
	Bet_Rot_Single12 = Bet_Rot_SingleStart + 12
	Bet_Rot_Single13 = Bet_Rot_SingleStart + 13
	Bet_Rot_Single14 = Bet_Rot_SingleStart + 14
	Bet_Rot_Single15 = Bet_Rot_SingleStart + 15
	Bet_Rot_Single16 = Bet_Rot_SingleStart + 16
	Bet_Rot_Single17 = Bet_Rot_SingleStart + 17
	Bet_Rot_Single18 = Bet_Rot_SingleStart + 18
	Bet_Rot_Single19 = Bet_Rot_SingleStart + 19
	Bet_Rot_Single20 = Bet_Rot_SingleStart + 20
	Bet_Rot_Single21 = Bet_Rot_SingleStart + 21
	Bet_Rot_Single22 = Bet_Rot_SingleStart + 22
	Bet_Rot_Single23 = Bet_Rot_SingleStart + 23
	Bet_Rot_Single24 = Bet_Rot_SingleStart + 24
	Bet_Rot_Single25 = Bet_Rot_SingleStart + 25
	Bet_Rot_Single26 = Bet_Rot_SingleStart + 26
	Bet_Rot_Single27 = Bet_Rot_SingleStart + 27
	Bet_Rot_Single28 = Bet_Rot_SingleStart + 28
	Bet_Rot_Single29 = Bet_Rot_SingleStart + 29
	Bet_Rot_Single30 = Bet_Rot_SingleStart + 30
	Bet_Rot_Single31 = Bet_Rot_SingleStart + 31
	Bet_Rot_Single32 = Bet_Rot_SingleStart + 32
	Bet_Rot_Single33 = Bet_Rot_SingleStart + 33
	Bet_Rot_Single34 = Bet_Rot_SingleStart + 34
	Bet_Rot_Single35 = Bet_Rot_SingleStart + 35
	Bet_Rot_Single36 = Bet_Rot_SingleStart + 36

	//双数(2区)  1赔17   ( rgnMain:1   rgnsub: 0 -- 59   压双数 )
	Bet_Rot_DoubleStart = Bet_Rot_Single36 + 1

	Bet_Rot_Double1or2   = Bet_Rot_DoubleStart + 0
	Bet_Rot_Double2or3   = Bet_Rot_DoubleStart + 1
	Bet_Rot_Double4or5   = Bet_Rot_DoubleStart + 2
	Bet_Rot_Double5or6   = Bet_Rot_DoubleStart + 3
	Bet_Rot_Double7or8   = Bet_Rot_DoubleStart + 4
	Bet_Rot_Double8or9   = Bet_Rot_DoubleStart + 5
	Bet_Rot_Double10or11 = Bet_Rot_DoubleStart + 6
	Bet_Rot_Double11or12 = Bet_Rot_DoubleStart + 7
	Bet_Rot_Double13or14 = Bet_Rot_DoubleStart + 8
	Bet_Rot_Double14or15 = Bet_Rot_DoubleStart + 9
	Bet_Rot_Double16or17 = Bet_Rot_DoubleStart + 10
	Bet_Rot_Double17or18 = Bet_Rot_DoubleStart + 11
	Bet_Rot_Double19or20 = Bet_Rot_DoubleStart + 12
	Bet_Rot_Double20or21 = Bet_Rot_DoubleStart + 13
	Bet_Rot_Double22or23 = Bet_Rot_DoubleStart + 14
	Bet_Rot_Double23or24 = Bet_Rot_DoubleStart + 15
	Bet_Rot_Double25or26 = Bet_Rot_DoubleStart + 16
	Bet_Rot_Double26or27 = Bet_Rot_DoubleStart + 17
	Bet_Rot_Double28or29 = Bet_Rot_DoubleStart + 18
	Bet_Rot_Double29or30 = Bet_Rot_DoubleStart + 19
	Bet_Rot_Double31or32 = Bet_Rot_DoubleStart + 20
	Bet_Rot_Double32or33 = Bet_Rot_DoubleStart + 21
	Bet_Rot_Double34or35 = Bet_Rot_DoubleStart + 22
	Bet_Rot_Double35or36 = Bet_Rot_DoubleStart + 23
	Bet_Rot_Double1or4   = Bet_Rot_DoubleStart + 24
	Bet_Rot_Double4or7   = Bet_Rot_DoubleStart + 25
	Bet_Rot_Double7or10  = Bet_Rot_DoubleStart + 26
	Bet_Rot_Double10or13 = Bet_Rot_DoubleStart + 27
	Bet_Rot_Double13or16 = Bet_Rot_DoubleStart + 28
	Bet_Rot_Double16or19 = Bet_Rot_DoubleStart + 29
	Bet_Rot_Double19or22 = Bet_Rot_DoubleStart + 30
	Bet_Rot_Double22or25 = Bet_Rot_DoubleStart + 31
	Bet_Rot_Double25or28 = Bet_Rot_DoubleStart + 32
	Bet_Rot_Double28or31 = Bet_Rot_DoubleStart + 33
	Bet_Rot_Double31or34 = Bet_Rot_DoubleStart + 34
	Bet_Rot_Double2or5   = Bet_Rot_DoubleStart + 35
	Bet_Rot_Double5or8   = Bet_Rot_DoubleStart + 36
	Bet_Rot_Double8or11  = Bet_Rot_DoubleStart + 37
	Bet_Rot_Double11or14 = Bet_Rot_DoubleStart + 38
	Bet_Rot_Double14or17 = Bet_Rot_DoubleStart + 39
	Bet_Rot_Double17or20 = Bet_Rot_DoubleStart + 40
	Bet_Rot_Double20or23 = Bet_Rot_DoubleStart + 41
	Bet_Rot_Double23or26 = Bet_Rot_DoubleStart + 42
	Bet_Rot_Double26or29 = Bet_Rot_DoubleStart + 43
	Bet_Rot_Double29or32 = Bet_Rot_DoubleStart + 44
	Bet_Rot_Double32or35 = Bet_Rot_DoubleStart + 45
	Bet_Rot_Double3or6   = Bet_Rot_DoubleStart + 46
	Bet_Rot_Double6or9   = Bet_Rot_DoubleStart + 47
	Bet_Rot_Double9or12  = Bet_Rot_DoubleStart + 48
	Bet_Rot_Double12or15 = Bet_Rot_DoubleStart + 49
	Bet_Rot_Double15or18 = Bet_Rot_DoubleStart + 50
	Bet_Rot_Double18or21 = Bet_Rot_DoubleStart + 51
	Bet_Rot_Double21or24 = Bet_Rot_DoubleStart + 52
	Bet_Rot_Double24or27 = Bet_Rot_DoubleStart + 53
	Bet_Rot_Double27or30 = Bet_Rot_DoubleStart + 54
	Bet_Rot_Double30or33 = Bet_Rot_DoubleStart + 55
	Bet_Rot_Double33or36 = Bet_Rot_DoubleStart + 56
	Bet_Rot_Double0or1   = Bet_Rot_DoubleStart + 57
	Bet_Rot_Double0or2   = Bet_Rot_DoubleStart + 58
	Bet_Rot_Double0or3   = Bet_Rot_DoubleStart + 59

	//三数(3区)  1赔11   ( rgnMain:2   rgnsub: 0 -- 13   压三数 )
	Bet_Rot_ThreeStart = Bet_Rot_Double0or3 + 1

	Bet_Rot_Three1to3    = Bet_Rot_ThreeStart + 0
	Bet_Rot_Three4to6    = Bet_Rot_ThreeStart + 1
	Bet_Rot_Three7to9    = Bet_Rot_ThreeStart + 2
	Bet_Rot_Three10to12  = Bet_Rot_ThreeStart + 3
	Bet_Rot_Three13to15  = Bet_Rot_ThreeStart + 4
	Bet_Rot_Three16to18  = Bet_Rot_ThreeStart + 5
	Bet_Rot_Three19to21  = Bet_Rot_ThreeStart + 6
	Bet_Rot_Three22to24  = Bet_Rot_ThreeStart + 7
	Bet_Rot_Three25to27  = Bet_Rot_ThreeStart + 8
	Bet_Rot_Three28to30  = Bet_Rot_ThreeStart + 9
	Bet_Rot_Three31to33  = Bet_Rot_ThreeStart + 10
	Bet_Rot_Three34to36  = Bet_Rot_ThreeStart + 11
	Bet_Rot_Three0to2    = Bet_Rot_ThreeStart + 12
	Bet_Rot_Three0or2or3 = Bet_Rot_ThreeStart + 13

	//四数(4区)  1赔8 ( rgnMain:3   rgnsub: 0 -- 22   压四个数的 )
	Bet_Rot_FourStart = Bet_Rot_Three0or2or3 + 1

	Bet_Rot_Four1or2or4or5     = Bet_Rot_FourStart + 0
	Bet_Rot_Four2or3or5or6     = Bet_Rot_FourStart + 1
	Bet_Rot_Four4or5or7or8     = Bet_Rot_FourStart + 2
	Bet_Rot_Four5or6or8or9     = Bet_Rot_FourStart + 3
	Bet_Rot_Four7or8or10or11   = Bet_Rot_FourStart + 4
	Bet_Rot_Four8or9or11or12   = Bet_Rot_FourStart + 5
	Bet_Rot_Four10or11or13or14 = Bet_Rot_FourStart + 6
	Bet_Rot_Four11or12or14or15 = Bet_Rot_FourStart + 7
	Bet_Rot_Four13or14or16or17 = Bet_Rot_FourStart + 8
	Bet_Rot_Four14or15or17or18 = Bet_Rot_FourStart + 9
	Bet_Rot_Four16or17or19or20 = Bet_Rot_FourStart + 10
	Bet_Rot_Four17or18r20or21  = Bet_Rot_FourStart + 11
	Bet_Rot_Four19or2or22or23  = Bet_Rot_FourStart + 12
	Bet_Rot_Four20or21or23or24 = Bet_Rot_FourStart + 13
	Bet_Rot_Four22or23or25or26 = Bet_Rot_FourStart + 14
	Bet_Rot_Four23or24or26or27 = Bet_Rot_FourStart + 15
	Bet_Rot_Four25or26or28or29 = Bet_Rot_FourStart + 16
	Bet_Rot_Four26or27or29or30 = Bet_Rot_FourStart + 17
	Bet_Rot_Four28or29or31or32 = Bet_Rot_FourStart + 18
	Bet_Rot_Four29or30or32or33 = Bet_Rot_FourStart + 19
	Bet_Rot_Four31or32or34or35 = Bet_Rot_FourStart + 20
	Bet_Rot_Four32or33or35or36 = Bet_Rot_FourStart + 21
	Bet_Rot_Four0or1or2or3     = Bet_Rot_FourStart + 22

	//六数(5区)	1赔5 ( rgnMain:4   rgnsub: 0 -- 10   //压6个数 )
	Bet_Rot_SixStart = Bet_Rot_Four0or1or2or3 + 1

	Bet_Rot_Six1to6   = Bet_Rot_SixStart + 0
	Bet_Rot_Six4to9   = Bet_Rot_SixStart + 1
	Bet_Rot_Six7to12  = Bet_Rot_SixStart + 2
	Bet_Rot_Six10to15 = Bet_Rot_SixStart + 3
	Bet_Rot_Six13to18 = Bet_Rot_SixStart + 4
	Bet_Rot_Six16to21 = Bet_Rot_SixStart + 5
	Bet_Rot_Six19to24 = Bet_Rot_SixStart + 6
	Bet_Rot_Six22to27 = Bet_Rot_SixStart + 7
	Bet_Rot_Six25to30 = Bet_Rot_SixStart + 8
	Bet_Rot_Six28to33 = Bet_Rot_SixStart + 9
	Bet_Rot_Six31to36 = Bet_Rot_SixStart + 10

	//12数(6区)	1赔2 ( rgnMain:5   rgnsub: 0 -- 5  压12个数  )
	Bet_Rot_TwelveStart = Bet_Rot_Six31to36 + 1

	Bet_Rot_Twelve1rd12 = Bet_Rot_TwelveStart + 0
	Bet_Rot_Twelve2nd12 = Bet_Rot_TwelveStart + 1
	Bet_Rot_Twelve3rd12 = Bet_Rot_TwelveStart + 2
	Bet_Rot_Twelve2to1a = Bet_Rot_TwelveStart + 3
	Bet_Rot_Twelve2to1b = Bet_Rot_TwelveStart + 4
	Bet_Rot_Twelve2to1c = Bet_Rot_TwelveStart + 5

	//红黑，小大，单双(7区)  1赔1 ( rgnMain:6   rgnsub: 0 -- 5  )
	Bet_Rot_EighteenStart = Bet_Rot_Twelve2to1c + 1

	Bet_Rot_EighteenRed    = Bet_Rot_EighteenStart + 0
	Bet_Rot_EighteenBlack  = Bet_Rot_EighteenStart + 1
	Bet_Rot_EighteenSmall  = Bet_Rot_EighteenStart + 2
	Bet_Rot_EighteenBig    = Bet_Rot_EighteenStart + 3
	Bet_Rot_EighteenSingle = Bet_Rot_EighteenStart + 4
	Bet_Rot_EighteenDouble = Bet_Rot_EighteenStart + 5

	//以下8-12五个区，是轮盘物理区下注，包括邻注/轮盘下角注/轮盘孤注/零旁注上角/轮上零旁，示意图如下：
	// 5  24 16 33 1  20 14 31 9  22 18 29 7  28 12 35
	// 10                                            3
	// 23  Tiers   \Orphelins   /Voisins       /Jeu0 26
	// 8                                             0
	// 30 11 36 13 27 6  34 17 25 2  21 4  19 15 32

	Bet_Rot_NeighbourTwoStart = Bet_Rot_EighteenDouble + 1

	//Neighbour邻注(8区)
	//说明：目前这版先实现每个数的左右两个邻居(NeighbourTwo)
	//下注：筹码需要是5的倍数，相当于下了5个单注
	//开奖：如果命中5个中的某个数字，赔率固定是1陪35

	Bet_Rot_NeighbourTwo0  = Bet_Rot_NeighbourTwoStart + 0
	Bet_Rot_NeighbourTwo32 = Bet_Rot_NeighbourTwoStart + 1
	Bet_Rot_NeighbourTwo15 = Bet_Rot_NeighbourTwoStart + 2
	Bet_Rot_NeighbourTwo19 = Bet_Rot_NeighbourTwoStart + 3
	Bet_Rot_NeighbourTwo4  = Bet_Rot_NeighbourTwoStart + 4
	Bet_Rot_NeighbourTwo21 = Bet_Rot_NeighbourTwoStart + 5
	Bet_Rot_NeighbourTwo2  = Bet_Rot_NeighbourTwoStart + 6
	Bet_Rot_NeighbourTwo25 = Bet_Rot_NeighbourTwoStart + 7
	Bet_Rot_NeighbourTwo17 = Bet_Rot_NeighbourTwoStart + 8
	Bet_Rot_NeighbourTwo34 = Bet_Rot_NeighbourTwoStart + 9
	Bet_Rot_NeighbourTwo6  = Bet_Rot_NeighbourTwoStart + 10
	Bet_Rot_NeighbourTwo27 = Bet_Rot_NeighbourTwoStart + 11
	Bet_Rot_NeighbourTwo13 = Bet_Rot_NeighbourTwoStart + 12
	Bet_Rot_NeighbourTwo36 = Bet_Rot_NeighbourTwoStart + 13
	Bet_Rot_NeighbourTwo11 = Bet_Rot_NeighbourTwoStart + 14
	Bet_Rot_NeighbourTwo30 = Bet_Rot_NeighbourTwoStart + 15
	Bet_Rot_NeighbourTwo8  = Bet_Rot_NeighbourTwoStart + 16
	Bet_Rot_NeighbourTwo23 = Bet_Rot_NeighbourTwoStart + 17
	Bet_Rot_NeighbourTwo10 = Bet_Rot_NeighbourTwoStart + 18
	Bet_Rot_NeighbourTwo5  = Bet_Rot_NeighbourTwoStart + 19
	Bet_Rot_NeighbourTwo24 = Bet_Rot_NeighbourTwoStart + 20
	Bet_Rot_NeighbourTwo16 = Bet_Rot_NeighbourTwoStart + 21
	Bet_Rot_NeighbourTwo33 = Bet_Rot_NeighbourTwoStart + 22
	Bet_Rot_NeighbourTwo1  = Bet_Rot_NeighbourTwoStart + 23
	Bet_Rot_NeighbourTwo20 = Bet_Rot_NeighbourTwoStart + 24
	Bet_Rot_NeighbourTwo14 = Bet_Rot_NeighbourTwoStart + 25
	Bet_Rot_NeighbourTwo31 = Bet_Rot_NeighbourTwoStart + 26
	Bet_Rot_NeighbourTwo9  = Bet_Rot_NeighbourTwoStart + 27
	Bet_Rot_NeighbourTwo22 = Bet_Rot_NeighbourTwoStart + 28
	Bet_Rot_NeighbourTwo18 = Bet_Rot_NeighbourTwoStart + 29
	Bet_Rot_NeighbourTwo29 = Bet_Rot_NeighbourTwoStart + 30
	Bet_Rot_NeighbourTwo7  = Bet_Rot_NeighbourTwoStart + 31
	Bet_Rot_NeighbourTwo28 = Bet_Rot_NeighbourTwoStart + 32
	Bet_Rot_NeighbourTwo12 = Bet_Rot_NeighbourTwoStart + 33
	Bet_Rot_NeighbourTwo35 = Bet_Rot_NeighbourTwoStart + 34
	Bet_Rot_NeighbourTwo3  = Bet_Rot_NeighbourTwoStart + 35
	Bet_Rot_NeighbourTwo26 = Bet_Rot_NeighbourTwoStart + 36

	//Tiers，轮盘下角注(9区)
	//说明：0对面的区域，包括30,11,36,13,27,33,16,24,5,10,23,8
	//下注：筹码需要是6的倍数，其中包括5,8及10,11及13,16及23,24及27,30及33,36六个双数注
	//开奖，如果命中其中的某个数字，赔率固定是1陪17
	Bet_Rot_Tiers = Bet_Rot_NeighbourTwo26 + 1

	//Orphelins，轮盘孤注(10区)
	//说明：轮盘侧边的6,34,17,9,31,14,20,1
	//下注：筹码需要是5的倍数，其中包括了1一个单注，和6,9及14,17及17,20及31,34四个双数注
	//开奖：如果开1则赔率是1陪35，如果不是1则赔率是1陪17
	Bet_Rot_Orphelins = Bet_Rot_NeighbourTwo26 + 2

	//Voisins，零旁注上角(11区)
	//说明：轮盘上0边上的25,2,21,4,19,15,32,0,26,3,35,12,28,7,29,18,22
	//下注：筹码需要是9的倍数，其中包括了0,2,3一个三数注(注意要两个筹码)、25,26,28,29一个四数注(注意要两个筹码)、4,7和12,15等5个双数注，加起来一共需要9个筹码
	//开奖：如果开0或2或3赔率是1陪11，如果开25或26或28或29赔率是1陪8，开其他数则赔率是1陪17
	Bet_Rot_Voisins = Bet_Rot_NeighbourTwo26 + 3

	//Jeu0，轮上零旁(12区)
	//说明：轮盘上0边上的15,32,0,26,3,35,12
	//下注：筹码需要是4的倍数，其中包括了26一个单注，0,3双数注，12,15双数注，32,35双数注
	//开奖：如果开26赔率是1陪35，如果不是26则是1陪17
	Bet_Rot_Jeu0 = Bet_Rot_NeighbourTwo26 + 4

	Bet_Rot_End = Bet_Rot_Jeu0

	/**************************************************************************/

	Bet_Sicbo_Start = 1500

	/*=============================================================================================
	// 骰子宝投注区域
	// 主区域代号： 大小单双：0   特定双数：1 特定三数：2 任意三数：3 总点数：4 一对组合：5   数字赌注：6
	//============================================================================================= */

	Bet_Sicbo_One_Big   = Bet_Sicbo_Start + 0 //大
	Bet_Sicbo_One_Small = Bet_Sicbo_Start + 1 //小
	Bet_Sicbo_One_Odd   = Bet_Sicbo_Start + 2 //奇数
	Bet_Sicbo_One_Even  = Bet_Sicbo_Start + 3 //偶数

	Bet_Sicbo_TwoNum_1And2 = Bet_Sicbo_Start + 4  //特定双数1和2
	Bet_Sicbo_TwoNum_1And3 = Bet_Sicbo_Start + 5  //特定双数1和3
	Bet_Sicbo_TwoNum_1And4 = Bet_Sicbo_Start + 6  //特定双数1和4
	Bet_Sicbo_TwoNum_1And5 = Bet_Sicbo_Start + 7  //特定双数1和5
	Bet_Sicbo_TwoNum_1And6 = Bet_Sicbo_Start + 8  //特定双数1和6
	Bet_Sicbo_TwoNum_2And3 = Bet_Sicbo_Start + 9  //特定双数2和3
	Bet_Sicbo_TwoNum_2And4 = Bet_Sicbo_Start + 10 //特定双数2和4
	Bet_Sicbo_TwoNum_2And5 = Bet_Sicbo_Start + 11 //特定双数2和5
	Bet_Sicbo_TwoNum_2And6 = Bet_Sicbo_Start + 12 //特定双数2和6
	Bet_Sicbo_TwoNum_3And4 = Bet_Sicbo_Start + 13 //特定双数3和4
	Bet_Sicbo_TwoNum_3And5 = Bet_Sicbo_Start + 14 //特定双数3和5
	Bet_Sicbo_TwoNum_3And6 = Bet_Sicbo_Start + 15 //特定双数3和6
	Bet_Sicbo_TwoNum_4And5 = Bet_Sicbo_Start + 16 //特定双数4和5
	Bet_Sicbo_TwoNum_4And6 = Bet_Sicbo_Start + 17 //特定双数4和6
	Bet_Sicbo_TwoNum_5And6 = Bet_Sicbo_Start + 18 //特定双数5和6

	Bet_Sicbo_ThreeSame_1All = Bet_Sicbo_Start + 19 //特定三数都是1
	Bet_Sicbo_ThreeSame_2All = Bet_Sicbo_Start + 20 //特定三数都是2
	Bet_Sicbo_ThreeSame_3All = Bet_Sicbo_Start + 21 //特定三数都是3
	Bet_Sicbo_ThreeSame_4All = Bet_Sicbo_Start + 22 //特定三数都是4
	Bet_Sicbo_ThreeSame_5All = Bet_Sicbo_Start + 23 //特定三数都是5
	Bet_Sicbo_ThreeSame_6All = Bet_Sicbo_Start + 24 //特定三数都是6

	Bet_Sicbo_ThreeSame_Any = Bet_Sicbo_Start + 25 //任意三数(全围）

	Bet_Sicbo_SumPoint_4  = Bet_Sicbo_Start + 26 //点数和4
	Bet_Sicbo_SumPoint_5  = Bet_Sicbo_Start + 27 //点数和5
	Bet_Sicbo_SumPoint_6  = Bet_Sicbo_Start + 28 //点数和6
	Bet_Sicbo_SumPoint_7  = Bet_Sicbo_Start + 29 //点数和7
	Bet_Sicbo_SumPoint_8  = Bet_Sicbo_Start + 30 //点数和8
	Bet_Sicbo_SumPoint_9  = Bet_Sicbo_Start + 31 //点数和9
	Bet_Sicbo_SumPoint_10 = Bet_Sicbo_Start + 32 //点数和10
	Bet_Sicbo_SumPoint_11 = Bet_Sicbo_Start + 33 //点数和11
	Bet_Sicbo_SumPoint_12 = Bet_Sicbo_Start + 34 //点数和12
	Bet_Sicbo_SumPoint_13 = Bet_Sicbo_Start + 35 //点数和13
	Bet_Sicbo_SumPoint_14 = Bet_Sicbo_Start + 36 //点数和14
	Bet_Sicbo_SumPoint_15 = Bet_Sicbo_Start + 37 //点数和15
	Bet_Sicbo_SumPoint_16 = Bet_Sicbo_Start + 38 //点数和16
	Bet_Sicbo_SumPoint_17 = Bet_Sicbo_Start + 39 //点数和17

	Bet_Sicbo_PairSame_1 = Bet_Sicbo_Start + 40 //一对组合点数1
	Bet_Sicbo_PairSame_2 = Bet_Sicbo_Start + 41 //一对组合点数2
	Bet_Sicbo_PairSame_3 = Bet_Sicbo_Start + 42 //一对组合点数3
	Bet_Sicbo_PairSame_4 = Bet_Sicbo_Start + 43 //一对组合点数4
	Bet_Sicbo_PairSame_5 = Bet_Sicbo_Start + 44 //一对组合点数5
	Bet_Sicbo_PairSame_6 = Bet_Sicbo_Start + 45 //一对组合点数6

	Bet_Sicbo_NumPoint_1 = Bet_Sicbo_Start + 46 //数字点数 1
	Bet_Sicbo_NumPoint_2 = Bet_Sicbo_Start + 47 //数字点数 2
	Bet_Sicbo_NumPoint_3 = Bet_Sicbo_Start + 48 //数字点数 3
	Bet_Sicbo_NumPoint_4 = Bet_Sicbo_Start + 49 //数字点数 4
	Bet_Sicbo_NumPoint_5 = Bet_Sicbo_Start + 50 //数字点数 5
	Bet_Sicbo_NumPoint_6 = Bet_Sicbo_Start + 51 //数字点数 6

	Bet_Sicbo_End = Bet_Sicbo_NumPoint_6

	/*=============================================================================================
	// 番摊投注区域
	// 主区域代号： 0:小大(0-1)，1:单双(0-1)，2:番(0--3),3:角(0--3),4:稔(0--11),5:三红/三门(0--3),6:正(0--3),7:通(0--11),
	// 开1,中奖区域ID:1600,1602,1604,1608,1611,1612,1613,1614,1615,1618,1621,1624,1625,1627
	// 开2,中奖区域ID:1600,1603,1605,1608,1609,1612,1615,1616,1617,1619,1622,1624,1625,1626
	// 开3,中奖区域ID:1601,1602,1606,1609,1610,1613,1616,1618,1619,1620,1623,1624,1626,1627
	// 开4,中奖区域ID:1601,1603,1607,1610,1611,1614,1617,1620,1621,1622,1623,1625,1626,1627
	//============================================================================================= */

	Bet_Ftan_Start = 1600

	//*******小大,单双,赔率(1:0.95)
	Bet_Ftan_Small_Big_And_Single_Double_Start = Bet_Ftan_Start
	Bet_Ftan_Small                             = Bet_Ftan_Small_Big_And_Single_Double_Start + 0
	Bet_Ftan_Big                               = Bet_Ftan_Small_Big_And_Single_Double_Start + 1
	Bet_Ftan_Single                            = Bet_Ftan_Small_Big_And_Single_Double_Start + 2
	Bet_Ftan_Double                            = Bet_Ftan_Small_Big_And_Single_Double_Start + 3

	Bet_Ftan_Small_Big_And_Single_Double_End = Bet_Ftan_Double

	//*******番,赔率(1:2.85)
	Bet_Ftan_Fan_Start = Bet_Ftan_Small_Big_And_Single_Double_End + 1

	Bet_Ftan_1Fan    = Bet_Ftan_Fan_Start + 0 //1番
	Bet_Ftan_2Fan    = Bet_Ftan_Fan_Start + 1 //2番
	Bet_Ftan_3Fan    = Bet_Ftan_Fan_Start + 2 //3番
	Bet_Ftan_4Fan    = Bet_Ftan_Fan_Start + 3 //4番
	Bet_Ftan_Fan_End = Bet_Ftan_4Fan

	//*******角,赔率(1:0.95)
	Bet_Ftan_Jiao_Start = Bet_Ftan_Fan_End + 1

	Bet_Ftan_12Jiao = Bet_Ftan_Jiao_Start + 0 //12角
	Bet_Ftan_23Jiao = Bet_Ftan_Jiao_Start + 1 //23角
	Bet_Ftan_34Jiao = Bet_Ftan_Jiao_Start + 2 //34角
	Bet_Ftan_41Jiao = Bet_Ftan_Jiao_Start + 3 //41角

	Bet_Ftan_Jiao_End = Bet_Ftan_41Jiao

	//*******念,赔率(1:1.9)
	Bet_Ftan_Nian_Start = Bet_Ftan_Jiao_End + 1
	Bet_Ftan_1Nian2     = Bet_Ftan_Nian_Start + 0 //1念2
	Bet_Ftan_1Nian3     = Bet_Ftan_Nian_Start + 1 //1念3
	Bet_Ftan_1Nian4     = Bet_Ftan_Nian_Start + 2 //1念4

	Bet_Ftan_2Nian1 = Bet_Ftan_Nian_Start + 3 //2念1
	Bet_Ftan_2Nian3 = Bet_Ftan_Nian_Start + 4 //2念3
	Bet_Ftan_2Nian4 = Bet_Ftan_Nian_Start + 5 //2念4

	Bet_Ftan_3Nian1 = Bet_Ftan_Nian_Start + 6 //3念1
	Bet_Ftan_3Nian2 = Bet_Ftan_Nian_Start + 7 //3念2
	Bet_Ftan_3Nian4 = Bet_Ftan_Nian_Start + 8 //3念4

	Bet_Ftan_4Nian1 = Bet_Ftan_Nian_Start + 9  //4念1
	Bet_Ftan_4Nian2 = Bet_Ftan_Nian_Start + 10 //4念2
	Bet_Ftan_4Nian3 = Bet_Ftan_Nian_Start + 11 //4念3

	Bet_Ftan_Nian_End = Bet_Ftan_4Nian3

	//*******三红/三门,赔率(1:0.31)
	Bet_Ftan_ThreeRed_Start = Bet_Ftan_Nian_End + 1
	Bet_Ftan_321ThreeRed    = Bet_Ftan_ThreeRed_Start + 0 //321三红
	Bet_Ftan_214ThreeRed    = Bet_Ftan_ThreeRed_Start + 1 //214三红
	Bet_Ftan_432ThreeRed    = Bet_Ftan_ThreeRed_Start + 2 //432三红
	Bet_Ftan_143ThreeRed    = Bet_Ftan_ThreeRed_Start + 3 //143三红

	Bet_Ftan_ThreeRed_End = Bet_Ftan_143ThreeRed

	Bet_Ftan_End = Bet_Ftan_ThreeRed_End

	/*=============================================================================================
	// 色碟投注区域
	// 主区域代号： 0:小大单双(0-4)，1:4白4红(0-1)，2:3白1红/3红1白(0--1),3:2白2红(0),4:4白或4红(0)
	//============================================================================================= */
	Bet_Sdie_Start = 1700

	Bet_Sdie_Single = Bet_Sdie_Start + 0 //单
	Bet_Sdie_Double = Bet_Sdie_Start + 1 //双
	Bet_Sdie_Small  = Bet_Sdie_Start + 2 //小
	Bet_Sdie_Big    = Bet_Sdie_Start + 3 //大

	Bet_Sdie_0Red_4White = Bet_Sdie_Start + 4 //0红4白
	Bet_Sdie_4Red_0White = Bet_Sdie_Start + 5 //4红

	Bet_Sdie_1Red_3White = Bet_Sdie_Start + 6 //1红3白
	Bet_Sdie_3Red_1White = Bet_Sdie_Start + 7 //3红1白

	Bet_Sdie_2Red_2White = Bet_Sdie_Start + 8 //2红2白

	Bet_Sdie_4White_Or_4Red = Bet_Sdie_Start + 9 //4白或4红
	Bet_Sdie_End            = Bet_Sdie_4White_Or_4Red

	/*=============================================================================================
	// 鱼虾蟹投注区域
	// 鱼，虾,葫芦,鹿,蟹，鸡，依次代表数字（1，2，3，4，5，6）
	// 主区域代号：0:下注单一点数,出现单骰，双骰，全骰(0-5)，1:其它下注主区域(日后可扩展,比如单双，大小等等）
	//============================================================================================= */
	Bet_Fsc_Start = 1800

	Bet_Fsc_Single_Point1 = Bet_Fsc_Start + 0 //1：鱼
	Bet_Fsc_Single_Point2 = Bet_Fsc_Start + 1 //2：虾
	Bet_Fsc_Single_Point3 = Bet_Fsc_Start + 2 //3：葫芦
	Bet_Fsc_Single_Point4 = Bet_Fsc_Start + 3 //4：鹿
	Bet_Fsc_Single_Point5 = Bet_Fsc_Start + 4 //5：蟹
	Bet_Fsc_Single_Point6 = Bet_Fsc_Start + 5 //5：鸡

	Bet_Fsc_End = Bet_Fsc_Single_Point6

	Bet_Area_End = Bet_Fsc_End + 1
)

type BetAreaDesc struct {
	AreadId int    //区域id
	Memo    string //区域的说明
}

var GlobalMapBetAreaDesc map[int]string //对个区域的备注说明

func initAreadIdCommnet() {

	betAreads := []BetAreaDesc{{Bet_Sicbo_One_Big, "Bet_Sicbo_One_Big"},
		{Bet_Sicbo_One_Big, "Bet_Sicbo_One_Small"},
	}

	GlobalMapBetAreaDesc = make(map[int]string)

	for i := 0; i < len(betAreads); i++ {

		id := betAreads[i].AreadId

		GlobalMapBetAreaDesc[id] = betAreads[i].Memo
	}

}

// 主取区域的说明，主要是用于测试用
func GetGlobalAreadIdComment(areadId int) string {

	memo, ok := GlobalMapBetAreaDesc[areadId]

	if ok == false {
		return ""
	}

	return memo
}
