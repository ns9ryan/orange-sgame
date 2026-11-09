package util

import (
	"fmt"
	"sync"
	"time"
)

var (
	counter int64
	mu      sync.Mutex // 保护 counter，防止并发冲突
)

func GenerateBetAccountId(deskno int, set int, round int) string {

	nowTime := time.Now()

	//年月日小时桌号 场次 局号

	strid := fmt.Sprintf("%02d%02d%02d%02d%02d%02d%02d  ", nowTime.Year()%1000, nowTime.Month(), nowTime.Day(), nowTime.Hour(), deskno, set, round)

	return strid
}

func GenerateRoundId(deskno int, set int, round int) string {

	nowTime := time.Now()

	//年月日小时桌号 场次 局号

	strid := fmt.Sprintf("OR%02d%02d%02d%02d%02d%02d%02d", nowTime.Year()%1000, nowTime.Month(), nowTime.Day(), nowTime.Hour(), deskno, set, round)

	return strid
}
