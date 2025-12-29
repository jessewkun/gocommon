package mysql

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// DateTime 自定义时间类型，用于在数据库中存储年-月-日 时:分:秒格式
type DateTime time.Time

// Value 实现 driver.Valuer 接口
func (t DateTime) Value() (driver.Value, error) {
	// 如果时间是零值，则存入数据库时为 NULL
	if time.Time(t).IsZero() {
		return nil, nil
	}
	return time.Time(t).Format("2006-01-02 15:04:05"), nil
}

// Scan 实现 sql.Scanner 接口
func (t *DateTime) Scan(value interface{}) error {
	if value == nil {
		*t = DateTime(time.Time{})
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*t = DateTime(v)
	case []byte:
		// 显式使用 time.Local，使时区行为明确
		parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", string(v), time.Local)
		if err != nil {
			return err
		}
		*t = DateTime(parsedTime)
	case string:
		// 显式使用 time.Local，使时区行为明确
		parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", v, time.Local)
		if err != nil {
			return err
		}
		*t = DateTime(parsedTime)
	default:
		return fmt.Errorf("cannot scan %T into DateTime", value)
	}
	return nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (t DateTime) MarshalJSON() ([]byte, error) {
	// 如果是零值，JSON 序列化为 null
	if time.Time(t).IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, time.Time(t).Format("2006-01-02 15:04:05"))), nil
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (t *DateTime) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" {
		*t = DateTime(time.Time{})
		return nil
	}
	if len(str) < 2 || str[0] != '"' || str[len(str)-1] != '"' {
		return fmt.Errorf("invalid DateTime format, expected a quoted string")
	}

	// 去除引号
	str = str[1 : len(str)-1]

	// 显式使用 time.Local，使时区行为明确
	parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", str, time.Local)
	if err != nil {
		return err
	}
	*t = DateTime(parsedTime)
	return nil
}

// String 实现 Stringer 接口
func (t DateTime) String() string {
	return time.Time(t).Format("2006-01-02 15:04:05")
}

// Format 将字符串转换为 DateTime 类型
func Format(datetime string) (DateTime, error) {
	// 显式使用 time.Local，使时区行为明确
	parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", datetime, time.Local)
	if err != nil {
		return DateTime(time.Time{}), err
	}
	return DateTime(parsedTime), nil
}
