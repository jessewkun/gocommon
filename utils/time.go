package utils

import (
	"time"
)

// IsDate 检查日期格式是否正确
//
// 返回是否正确
func IsDate(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}

// Today 获取当前日期
//
// 返回当前日期
func Today() string {
	return time.Now().Format("2006-01-02")
}

// Now 获取当前时间
//
// 返回当前时间
func Now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// NowTimeStamp 获取当前时间戳
//
// 返回当前时间戳
func NowTimeStamp() int64 {
	return time.Now().Unix()
}

// TimestampToDate 将时间戳转换为日期
//
// 返回日期
func TimestampToDate(timestamp int64) string {
	return time.Unix(timestamp, 0).Format("2006-01-02")
}

// DatetimeToTime 将字符串转换为时间
//
// 返回时间
func DatetimeToTime(datetime string) time.Time {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", datetime, loc)
	return t
}

// TimeDifference 计算两个时间的时间差
//
// 返回两个时间的时间差（秒）
func TimeDifference(t1, t2 string) int64 {
	return DatetimeToTime(t1).Unix() - DatetimeToTime(t2).Unix()
}

// GetDayTimeRange 获取一天的时间范围
//
// 返回一天的开始时间和结束时间
func GetDayTimeRange(t time.Time) (start, end time.Time) {
	start = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	end = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
	return
}
