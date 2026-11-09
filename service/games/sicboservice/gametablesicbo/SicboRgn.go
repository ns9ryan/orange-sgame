package gameTableSicbo

//=============================================================================================
// 骰子宝投注区域
// 主区域代号： 大小单双：0   特定双数：1 特定三数：2 任意三数：3 总点数：4 一对组合：5   数字赌注：6
//=============================================================================================

const (
	OneIdNum          = 4  //大，小，奇，偶
	TwoSpecIdNum      = 15 //二个特别点数一样
	ThreeSameIdNum    = 6  //同时三个一样点数
	ThreeAnySameIdNum = 1  //任意三个一样点数
	SumTotalIdNum     = 14 //总点数
	PairIdNum         = 6  //对子
	NumberIdNum       = 6

	MaxIdNum = OneIdNum + TwoSpecIdNum + ThreeSameIdNum + ThreeAnySameIdNum + SumTotalIdNum + PairIdNum + NumberIdNum
)

const TotalAnswer = 56 //56种开将组合
const TwoAnswer = TwoSpecIdNum

var SicboAnswerInds [TotalAnswer][3]int  //开奖数组表
var SicboAnswerTwoInds [TwoAnswer][2]int //二个点数组合开奖列表

func init() {

	//开奖的组合
	tempArray := [TotalAnswer][3]int{
		{1, 1, 1}, {1, 1, 2}, {1, 1, 3}, {1, 2, 2}, {1, 1, 4}, {1, 2, 3}, {2, 2, 2}, {1, 1, 5}, {1, 2, 4}, {1, 3, 3}, {2, 2, 3}, {1, 1, 6}, {1, 2, 5}, {1, 3, 4},
		{2, 2, 4}, {2, 3, 3}, {1, 2, 6}, {1, 3, 5}, {1, 4, 4}, {2, 3, 4}, {2, 2, 5}, {3, 3, 3}, {1, 3, 6}, {1, 4, 5}, {2, 2, 6}, {2, 3, 5}, {2, 4, 4}, {3, 3, 4}, /*14*/
		{1, 4, 6}, {1, 5, 5}, {2, 3, 6}, {2, 4, 5}, {3, 3, 5}, {3, 4, 4}, {1, 5, 6}, {2, 4, 6}, {2, 5, 5}, {3, 3, 6}, {3, 4, 5}, {4, 4, 4}, {1, 6, 6}, {2, 5, 6}, /*28*/
		{3, 4, 6}, {3, 5, 5}, {4, 4, 5}, {2, 6, 6}, {3, 5, 6}, {4, 4, 6}, {4, 5, 5}, {3, 6, 6}, {4, 5, 6}, {5, 5, 5}, {4, 6, 6}, {5, 5, 6}, {5, 6, 6}, {6, 6, 6}, /*42*/
	}

	twoArray := [TwoSpecIdNum][2]int{{1, 2}, {1, 3}, {1, 4}, {1, 5}, {1, 6}, {2, 3}, {2, 4}, {2, 5}, {2, 6}, {3, 4}, {3, 5}, {3, 6}, {4, 5}, {4, 6}, {5, 6}}

	SicboAnswerInds = tempArray
	SicboAnswerTwoInds = twoArray

}

// 输入点数找到对点的索引
func SicboFindPointsToAnswerIndex(point1, point2, point3 int) int {

	answer := -1

	for i := 0; i < TotalAnswer; i++ {

		points := []int{point1, point2, point3}

		num1 := point1 + point2 + point3
		num2 := SicboAnswerInds[i][0] + SicboAnswerInds[i][1] + SicboAnswerInds[i][2]

		if num1 != num2 {
			continue
		}

		for j := 0; j < 3; j++ {

			point := SicboAnswerInds[i][j]

			for index, value := range points {
				//若找到则置0
				if value == point {
					points[index] = 0
					break
				}
			}
		}

		num := points[0] + points[1] + points[2]

		//找到了
		if num == 0 {
			answer = i
			break
		}
	}

	return answer
}

// 投注区的二维结构体
type RotRgn struct {
	RgnMain int
	RgnSub  int
}

func (rot *RotRgn) IsSame() {

}
