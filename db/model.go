package db

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// DateTime 自定义时间类型，用于在数据库中存储年-月-日 时:分:秒格式
type DateTime time.Time

// Value 实现 driver.Valuer 接口
func (t DateTime) Value() (driver.Value, error) {
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
	case string:
		parsedTime, err := time.Parse("2006-01-02 15:04:05", v)
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
	return []byte(fmt.Sprintf(`"%s"`, time.Time(t).Format("2006-01-02 15:04:05"))), nil
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (t *DateTime) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" {
		*t = DateTime(time.Time{})
		return nil
	}

	// 去除引号
	str = str[1 : len(str)-1]

	parsedTime, err := time.Parse("2006-01-02 15:04:05", str)
	if err != nil {
		return err
	}
	*t = DateTime(parsedTime)
	return nil
}

// BaseModel 定义基础字段，方便所有业务模型继承
type BaseModel struct {
	ID        uint     `gorm:"primarykey" json:"id"`
	CreatedAt DateTime `gorm:"type:datetime" json:"created_at"`
	UpdatedAt DateTime `gorm:"type:datetime" json:"updated_at"`
}
