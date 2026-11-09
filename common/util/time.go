package util

import "time"

// TimestampToDatetime 将 Unix 时间戳转为指定时区的时间
func TimestampToDatetime(timestamp int64) string {
	loc, _ := time.LoadLocation(DefaultLocation) // 默认中国的时区（UTC+8）
	return time.Unix(timestamp, 0).In(loc).Format("2006-01-02 15:04:05")
}
