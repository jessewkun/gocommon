package mysql

// BaseModel 定义基础字段，方便所有业务模型继承
type BaseModel struct {
	ID         int      `gorm:"primarykey" json:"id"`
	CreatedAt  DateTime `gorm:"type:datetime" json:"created_at"`
	ModifiedAt DateTime `gorm:"type:datetime" json:"modified_at"`
}
