package utils

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type LocalTime time.Time

const TIME_FORMAT = "2006-01-02 15:04:05"
const DATE_FORMAT = "2006-01-02"
const TIMEZONE = "Asia/Shanghai"

// MarshalJSON 将时间转换为字符串
//
// 返回字符串
func (t LocalTime) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(TIME_FORMAT)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, TIME_FORMAT)
	b = append(b, '"')
	return b, nil
}

// UnmarshalJSON 将字符串转换为时间
//
// 返回时间
func (t *LocalTime) UnmarshalJSON(data []byte) (err error) {
	now, err := time.ParseInLocation(`"`+TIME_FORMAT+`"`, string(data), time.Local)
	*t = LocalTime(now)
	return
}

// String 将时间转换为字符串
//
// 返回字符串
func (t LocalTime) String() string {
	return time.Time(t).Format(TIME_FORMAT)
}

// local 将时间转换为本地时间
//
// 返回本地时间
func (t LocalTime) local() time.Time {
	loc, _ := time.LoadLocation(TIMEZONE)
	return time.Time(t).In(loc)
}

// Value 将时间转换为时间戳
//
// 返回时间戳
func (t LocalTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	var ti = time.Time(t)
	if ti.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return ti, nil
}

// Date 将时间转换为日期
//
// 返回日期
func (t LocalTime) Date() string {
	return time.Time(t).Format(DATE_FORMAT)
}

// Scan 将时间转换为时间戳
//
// 返回时间戳
func (t *LocalTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = LocalTime(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

func (t LocalTime) Format(format string) string {
	return time.Time(t).Format(format)
}

// IsDate 检查日期格式是否正确
//
// 返回是否正确
func IsDate(date string) bool {
	_, err := time.Parse(DATE_FORMAT, date)
	return err == nil
}

// Today 获取当前日期
//
// 返回当前日期
func Today() string {
	return time.Now().Format(DATE_FORMAT)
}

// Now 获取当前时间
//
// 返回当前时间
func Now() string {
	return time.Now().Format(TIME_FORMAT)
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
	return time.Unix(timestamp, 0).Format(DATE_FORMAT)
}

// DatetimeToTime 将字符串转换为时间
//
// 返回时间
func DatetimeToTime(datetime string) time.Time {
	loc, _ := time.LoadLocation(TIMEZONE)
	t, _ := time.ParseInLocation(TIME_FORMAT, datetime, loc)
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
