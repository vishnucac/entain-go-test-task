package models

import (
	"entain-task/internal/db"
	"errors"
	"math/big"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type User struct {
	ID      uint64  `gorm:"primaryKey"`
	Balance float64 `gorm:"type:numeric(20,2);not null;default:0"`
}

type Transaction struct {
	ID            uint64 `gorm:"primaryKey;autoIncrement"`
	TransactionID string `gorm:"uniqueIndex;size:255;not null"`
	UserID        uint64 `gorm:"index;not null"`
	State         string `gorm:"type:varchar(10);not null"`
	Amount        float64
}

var ErrNegativeBalance = errors.New("balance cannot be negative")
var ErrDuplicateTransaction = errors.New("transaction already processed")

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Transaction{})
}

// UpdateUserBalance method updates user balance with transaction
func UpdateUserBalance(userID uint64, transactionID, state, amountStr string) error {
	amount, err := parseAmount(amountStr)
	if err != nil {
		return err
	}

	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	// Check if transactionId already exists
	var existingTx Transaction
	if err := tx.Where("transaction_id = ?", transactionID).First(&existingTx).Error; err == nil {
		tx.Rollback()
		return ErrDuplicateTransaction
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return err
	}

	var user User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
		tx.Rollback()
		return err
	}

	newBalance := user.Balance
	switch strings.ToLower(state) {
	case "win":
		newBalance += amount
	case "lose":
		newBalance -= amount
	default:
		tx.Rollback()
		return errors.New("invalid state")
	}

	// Account balance can't be in a negative value.
	if newBalance < 0 {
		tx.Rollback()
		return ErrNegativeBalance
	}

	// Update user balance
	if err := tx.Model(&user).Update("balance", newBalance).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Create transaction record
	newTx := Transaction{
		TransactionID: transactionID,
		UserID:        userID,
		State:         state,
		Amount:        amount,
	}
	if err := tx.Create(&newTx).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func parseAmount(s string) (float64, error) {
	f := new(big.Float)
	if _, ok := f.SetString(s); !ok {
		return 0, errors.New("invalid amount format")
	}
	// Round to 2 decimals:
	f.SetPrec(64)
	f64, _ := f.Float64()
	return roundTo2Decimals(f64), nil
}

func roundTo2Decimals(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
