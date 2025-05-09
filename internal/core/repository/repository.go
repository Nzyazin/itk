package repository

import (
	"context"

	"github.com/Nzyazin/itk/internal/core/models"
	"github.com/google/uuid"
)

type WalletRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error)
	GetCurrencyByCode(ctx context.Context, code string) (*models.Currency, error)
    ExecuteTxWithRetry(ctx context.Context, walletID uuid.UUID, amount int64, opType models.OperationType) (int64, error)
}
