package utils

import (
	"testing"
	"time"
)

func TestIsDate(t *testing.T) {
	if !IsDate("2024-06-19") {
		t.Error("IsDate 应该返回 true")
	}
	if IsDate("2024-13-01") {
		t.Error("IsDate 应该返回 false")
	}
	if IsDate("abc") {
		t.Error("IsDate 应该返回 false")
	}
}

func TestToday(t *testing.T) {
	today := Today()
	if !IsDate(today) {
		t.Errorf("Today 返回值不是有效日期: %s", today)
	}
}

func TestNow(t *testing.T) {
	now := Now()
	if len(now) != 19 {
		t.Errorf("Now 返回值长度不对: %s", now)
	}
}

func TestNowTimeStamp(t *testing.T) {
	ts := NowTimeStamp()
	if ts <= 0 {
		t.Errorf("NowTimeStamp 返回值不正确: %d", ts)
	}
}

func TestTimestampToDate(t *testing.T) {
	ts := int64(1718803200) // 2024-06-19 00:00:00
	date := TimestampToDate(ts)
	if date != "2024-06-19" {
		t.Errorf("TimestampToDate 返回值不正确: %s", date)
	}
}

func TestDatetimeToTime(t *testing.T) {
	dt := "2024-06-19 12:34:56"
	tm := DatetimeToTime(dt)
	if tm.Year() != 2024 || tm.Month() != 6 || tm.Day() != 19 || tm.Hour() != 12 || tm.Minute() != 34 || tm.Second() != 56 {
		t.Errorf("DatetimeToTime 解析错误: %v", tm)
	}
}

func TestTimeDifference(t *testing.T) {
	t1 := "2024-06-19 12:00:00"
	t2 := "2024-06-19 11:00:00"
	diff := TimeDifference(t1, t2)
	if diff != 3600 {
		t.Errorf("TimeDifference 结果不正确: %d", diff)
	}
}

func TestGetDayTimeRange(t *testing.T) {
	tm := time.Date(2024, 6, 19, 15, 30, 0, 0, time.Local)
	start, end := GetDayTimeRange(tm)
	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Errorf("GetDayTimeRange start 错误: %v", start)
	}
	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Errorf("GetDayTimeRange end 错误: %v", end)
	}
}
