package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Wallet представляет модель кошелька
type Wallet struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Balance   int64   `json:"balance" db:"balance"` // в копейках
	CurrencyCode  string    `json:"currency" db:"currency"` // ISO 4217: "USD", "RUB"
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// OperationType определяет тип операции с кошельком
type OperationType string

const (
	// OperationDeposit - пополнение кошелька
	OperationDeposit OperationType = "DEPOSIT"
	// OperationWithdraw - снятие средств с кошелька
	OperationWithdraw OperationType = "WITHDRAW"
)

// WalletOperation представляет запрос на операцию с кошельком
type WalletOperation struct {
	WalletID      uuid.UUID     `json:"walletId"`
	OperationType OperationType `json:"operationType"`
	Amount        string       `json:"amount"`
	DecimalAmount decimal.Decimal `json:"-"`
}
