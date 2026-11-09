package gameTableFtan

/*=============================================================================================
// 番摊投注区域
// 主区域代号： 0:小大单双(0-3)，1:番(0--3),2:角(0--3),3:念(0--11),4:三红/三门(0--3)
//============================================================================================= */

const (
	SmallBigAndSingleDoubleId = iota
	FanId
	JiaoId
	NianId
	ThreeRedId
)

const (
	SmallBigAndSingleDoubleIdNum = 4  //小大,单双
	FanIdNum                     = 4  //番
	JiaoIdNum                    = 4  //角
	NianIdNum                    = 12 //念
	ThreeRedIdNum                = 4  //三红
	MaxIdNum                     = SmallBigAndSingleDoubleIdNum + FanIdNum + JiaoIdNum + NianIdNum + ThreeRedIdNum
)

const TotalAnswer = 4 //4种开奖结果
var FanMember [4]int
var JiaoMember [4][2]int
var NianWinMember [12]int
var NianTieMember [12]int
var ThreeRedMember [4][3]int

var SmallBigMemberAndSingleDouble [4][2]int

func init() {

	//小大(0-1),单双 (2-3)
	SmallBigMemberAndSingleDouble = [4][2]int{{1, 2}, {3, 4}, {1, 3}, {2, 4}}

	//番
	FanMember = [4]int{1, 2, 3, 4}

	//角
	JiaoMember = [4][2]int{{1, 2}, {2, 3}, {3, 4}, {4, 1}}

	//念【NianWinMember[0]+NianTieMember[0]:1念4】....以此类推
	NianWinMember = [12]int{1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4}
	NianTieMember = [12]int{2, 3, 4, 1, 3, 4, 1, 2, 4, 1, 2, 3}

	//三红
	ThreeRedMember = [4][3]int{{1, 2, 3}, {1, 2, 4}, {2, 3, 4}, {1, 3, 4}}

}
