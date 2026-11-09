package gameTableSdie

/*=============================================================================================
// 色碟投注区域
// 主区域代号： 0:单双小大(0-3)，1:0红4白，4红0白(0-1)，2:1红3白,3红1白(0-1)，3:2红2白(0)，4:4红或4白(0)
//============================================================================================= */

const (
	SingleDoubleAndSmallBigId = iota
	FourWhiteFourRedId
	ThreeWhiteThreeRedId
	TwoWhiteTwoRedId
	FourWhiteOrFourRedId
)

const (
	SingleDoubleAndSmallBigNum = 4 //单,双,小,大
	FourWhiteFourRedIdNum      = 2 //0红4白，4红0白
	ThreeWhiteThreeRedIdNum    = 2 //1红3白，3红1白
	TwoWhiteTwoRedIdNum        = 1 //2红2白
	FourWhiteOrFourRedIdNum    = 1 //0红4白或4红0白
	MaxIdNum                   = SingleDoubleAndSmallBigNum + FourWhiteFourRedIdNum + ThreeWhiteThreeRedIdNum + TwoWhiteTwoRedIdNum + FourWhiteOrFourRedIdNum
)

const TotalAnswer = 5 //5种开奖结果

var SingleDoubleAndSmallBigMember [][]int

var FourWhiteFourRedMember [2]int

var ThreeWhiteThreeRedMember [2]int

var TwoWhiteTwoRedMember int

var FourWhiteOrFourRedMember [2]int

func init() {

	//单双(0-1),小大(2-3)
	SingleDoubleAndSmallBigMember = [][]int{{1, 3}, {0, 2, 4}, {0, 1}, {3, 4}}

	//0红4白,4红0白
	FourWhiteFourRedMember = [2]int{0, 4}

	//1红3白,3红1白
	ThreeWhiteThreeRedMember = [2]int{1, 3}

	//2红2白
	TwoWhiteTwoRedMember = 2

	//4红或4白
	FourWhiteOrFourRedMember = [2]int{0, 4}

}
