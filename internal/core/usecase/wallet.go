package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Nzyazin/itk/internal/core/logger"
	"github.com/Nzyazin/itk/internal/core/models"
	"github.com/Nzyazin/itk/internal/core/repository"
	"github.com/shopspring/decimal"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WalletUsecase interface {
	OperateWallet(ctx context.Context, op *models.WalletOperation) error
}

type walletUsecase struct {
	repo repository.WalletRepository
	log  logger.Logger
}

func (uc *walletUsecase) OperateWallet(ctx context.Context, op *models.WalletOperation) error {
    uc.logStart(op)
    
    wallet, err := uc.getWallet(ctx, op)
    if err != nil {
        return err
    }

    currency, err := uc.getCurrency(ctx, wallet)
    if err != nil {
        return err
    }

    amount, err := uc.convertAmount(op.Amount, currency)
    if err != nil {
        return err
    }

    amount, err = uc.checkBalance(wallet, amount, op.OperationType)
    if err != nil {
        return err
    }

    if err := uc.updateBalance(ctx, wallet.ID, amount, op.OperationType); err != nil {
        return err
    }

    return uc.createTransaction(ctx, wallet, amount, op)
}

func (uc *walletUsecase) logStart(op *models.WalletOperation) {
    uc.log.Info("Starting operation",
        zap.String("wallet_id", op.WalletID.String()),
        zap.String("type", string(op.OperationType)),
        zap.String("amount", op.Amount))
}

func (uc *walletUsecase) getWallet(ctx context.Context, op *models.WalletOperation) (*models.Wallet, error) {
    wallet, err := uc.repo.GetByID(ctx, op.WalletID)
    if err != nil {
        uc.log.Error("Wallet lookup failed", zap.Error(err), zap.Any("op", op))
        return nil, fmt.Errorf("get wallet: %w", err)
    }
    return wallet, nil
}

func (uc *walletUsecase) getCurrency(ctx context.Context, wallet *models.Wallet) (*models.Currency, error) {
    currency, err := uc.repo.GetCurrencyByCode(ctx, wallet.CurrencyCode)
    if err != nil {
        uc.log.Error("Currency error", 
            zap.Error(err),
            zap.String("code", wallet.CurrencyCode))
        return nil, fmt.Errorf("get currency: %w", err)
    }
    return currency, nil
}

func (uc *walletUsecase) convertAmount(input string, currency *models.Currency) (int64, error) {
    normalized := strings.ReplaceAll(input, ",", ".")
    amount, err := decimal.NewFromString(normalized)
    if err != nil {
        uc.log.Error("Amount conversion error",
            zap.String("input", input),
            zap.Error(err))
        return 0, fmt.Errorf("convert amount: %w", err)
    }
    return amount.Mul(decimal.NewFromInt(int64(currency.MinorUnits))).IntPart(), nil
}

func (uc *walletUsecase) checkBalance(wallet *models.Wallet, amount int64, opType models.OperationType) (int64, error) {
    if opType == models.OperationWithdraw && wallet.Balance < amount {
        uc.log.Warn("Insufficient funds",
            zap.Int64("balance", wallet.Balance),
            zap.Int64("requested", amount))
        return 0, errors.New("insufficient funds")
    }
    return amount, nil
}

func (uc *walletUsecase) updateBalance(ctx context.Context, walletID uuid.UUID, amount int64, opType models.OperationType) error {
    if opType == models.OperationWithdraw {
        amount = -amount
    }

    if err := uc.repo.UpdateBalance(ctx, walletID, amount, opType); err != nil {
        uc.log.Error("Balance update failed",
            zap.Error(err),
            zap.Int64("delta", amount))
        return fmt.Errorf("update balance: %w", err)
    }
    return nil
}

func (uc *walletUsecase) createTransaction(ctx context.Context, wallet *models.Wallet, amount int64, op *models.WalletOperation) error {
    tx := &models.Transaction{
        WalletID:      wallet.ID,
        OperationType: op.OperationType,
        Amount:        amount,
        Status:        "COMPLETED",
    }

    if err := uc.repo.CreateTransaction(ctx, tx); err != nil {
        uc.log.Error("Transaction failed",
            zap.Error(err),
            zap.Any("tx", tx))
        return fmt.Errorf("create transaction: %w", err)
    }

    uc.log.Info("Operation success",
        zap.String("tx_id", tx.ID.String()),
        zap.Int64("new_balance", wallet.Balance+amount))
    return nil
}
