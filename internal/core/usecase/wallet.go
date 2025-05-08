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
	"go.uber.org/zap"
)

type WalletUsecase interface {
	OperateWallet(ctx context.Context, op *models.WalletOperation) error
}

type walletUsecase struct {
	repo repository.WalletRepository
	log  logger.Logger
}

func (uc *walletUsecase) OperateWallet(ctx context.Context, op *models.WalletOperation) (int64, error) {
    uc.logStart(op)
    
    wallet, err := uc.getWallet(ctx, op)
    if err != nil {
        return 0, err
    }

    currency, err := uc.getCurrency(ctx, wallet)
    if err != nil {
        return 0, err
    }

    amount, err := uc.convertAmount(op.Amount, currency)
    if err != nil {
        return 0, err
    }

    amount, err = uc.checkBalance(wallet, amount, op.OperationType)
    if err != nil {
        return 0, err
    }

    newBalance, err := uc.repo.ExecuteTx(ctx, wallet.ID, amount, op.OperationType)
    if err != nil {
        return 0, err
    }

    return newBalance, nil
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

