package models

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID            uuid.UUID     `json:"id" db:"id"`
	WalletID      uuid.UUID     `json:"wallet_id" db:"wallet_id"`
	OperationType OperationType `json:"operation_type" db:"operation_type"`
	Amount        int64         `json:"amount" db:"amount"`
	Status        string        `json:"status" db:"status"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
}