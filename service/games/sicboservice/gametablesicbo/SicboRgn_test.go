package gameTableSicbo

import (
	"fmt"
	"testing"
)

func TestAdd(t *testing.T) {

	result := SicboFindPointsToAnswerIndex(6, 6, 6)

	if result == -1 {

		fmt.Println("SicboFindPointsToAnswerIndex fail")
		return
	}

	fmt.Printf("SicboFindPointsToAnswerIndex success index: [%d ]  \n", result)

}

func TestBetAread(t *testing.T) {

	pp := PlayerBet{}

	for i := 0; i < 55; i++ {

		pp.ThreeSameIdWin(i, 100, 5)
	}

}
