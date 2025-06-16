package db

import "gorm.io/gorm"

type Transaction struct {
	db *gorm.DB
	tx *gorm.DB
}

func NewTransaction(db *gorm.DB) *Transaction {
	return &Transaction{
		db: db,
		tx: db.Begin(),
	}
}

func (t *Transaction) Commit() error {
	return t.tx.Commit().Error
}

func (t *Transaction) Rollback() error {
	return t.tx.Rollback().Error
}
