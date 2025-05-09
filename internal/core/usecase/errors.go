package usecase

import "errors"

// Определение ошибок сервиса
var (
	ErrInvalidAmount      = errors.New("amount must be positive")
	ErrInvalidOperationType = errors.New("invalid operation type")
	ErrWalletNotFound     = errors.New("wallet not found")
	ErrInsufficientFunds  = errors.New("insufficient funds")
)
