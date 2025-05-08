package repository

import (
	"context"

	"github.com/Nzyazin/itk/internal/core/models"
	"github.com/google/uuid"
)

type WalletRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error)
	GetCurrencyByCode(ctx context.Context, code string) (*models.Currency, error)
	ExecuteTx(ctx context.Context, walletID uuid.UUID, amount int64, operationType models.OperationType) (newBalance int64, err error)
}
