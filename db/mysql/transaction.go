package mysql

import (
	"errors"

	"gorm.io/gorm"
)

type Transaction struct {
	db *gorm.DB
	tx *gorm.DB
}

func NewTransaction(db *gorm.DB) *Transaction {
	if db == nil {
		return &Transaction{
			db: nil,
			tx: nil,
		}
	}

	return &Transaction{
		db: db,
		tx: db.Begin(),
	}
}

func (t *Transaction) Commit() error {
	if t.tx == nil {
		return errors.New("transaction is nil, cannot commit")
	}
	return t.tx.Commit().Error
}

func (t *Transaction) Rollback() error {
	if t.tx == nil {
		return errors.New("transaction is nil, cannot rollback")
	}
	return t.tx.Rollback().Error
}
