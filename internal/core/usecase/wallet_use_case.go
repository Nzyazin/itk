package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/Nzyazin/itk/internal/core/logger"
	"github.com/Nzyazin/itk/internal/core/models"
	"github.com/Nzyazin/itk/internal/core/repository"
	"github.com/shopspring/decimal"
)

type WalletUsecase interface {
	OperateWallet(ctx context.Context, op models.WalletOperation) (decimal.Decimal, error)
}

type walletUsecase struct {
	repo repository.WalletRepository
	log  logger.Logger
}

func NewWalletUsecase(repo repository.WalletRepository, log logger.Logger) WalletUsecase {
	return &walletUsecase{repo: repo, log: log}
}

func (uc *walletUsecase) OperateWallet(ctx context.Context, op models.WalletOperation) (decimal.Decimal, error) {
    uc.logStart(op)
    
    wallet, err := uc.getWallet(ctx, op)
    if err != nil {
        return decimal.Zero, err
    }

    currency, err := uc.getCurrency(ctx, wallet)
    if err != nil {
        return decimal.Zero, err
    }

    amount, err := uc.convertAmountToMinorUnits(op.Amount, currency)
    if err != nil {
        return decimal.Zero, err
    }

    amount, err = uc.checkBalance(wallet, amount, op.OperationType)
    if err != nil {
        return decimal.Zero, err
    }

    newBalance, err := uc.repo.ExecuteTx(ctx, wallet.ID, amount, op.OperationType)
    if err != nil {
        return decimal.Zero, err
    }

    newBalanceStr, err := uc.convertAmountFromMinorUnits(newBalance, currency)
    if err != nil {
        return decimal.Zero, err
    }

    return newBalanceStr, nil
}

func (uc *walletUsecase) logStart(op models.WalletOperation) {
    uc.log.Info("Starting operation",
        logger.StringField("wallet_id", op.WalletID.String()),
        logger.StringField("type", string(op.OperationType)),
        logger.StringField("amount", op.Amount))
}

func (uc *walletUsecase) getWallet(ctx context.Context, op models.WalletOperation) (*models.Wallet, error) {
    wallet, err := uc.repo.GetByID(ctx, op.WalletID)
    if err != nil {
        uc.log.Error("Wallet lookup failed", 
            logger.ErrorField("error", err),
            logger.AnyField("op", op))
        return nil, fmt.Errorf("get wallet: %w", err)
    }
    return wallet, nil
}

func (uc *walletUsecase) getCurrency(ctx context.Context, wallet *models.Wallet) (*models.Currency, error) {
    currency, err := uc.repo.GetCurrencyByCode(ctx, wallet.CurrencyCode)
    if err != nil {
        uc.log.Error("Currency error", 
            logger.ErrorField("error", err),
            logger.StringField("code", wallet.CurrencyCode))
        return nil, fmt.Errorf("get currency: %w", err)
    }
    return currency, nil
}

func (uc *walletUsecase) convertAmountToMinorUnits(amountStr string, currency *models.Currency) (int64, error) {
    normalAmount := strings.ReplaceAll(amountStr, ",", ".")
    amount, err := decimal.NewFromString(normalAmount)
    if err != nil {
        uc.log.Error("Amount conversion error",
            logger.StringField("input", amountStr),
            logger.ErrorField("error", err))
        return 0, fmt.Errorf("convert amount: %w", err)
    }
    multiplier := decimal.NewFromInt(10).Pow(decimal.NewFromInt(currency.MinorUnits))
    return amount.Mul(multiplier).IntPart(), nil
}

func (uc *walletUsecase) convertAmountFromMinorUnits(minorUnits int64, currency *models.Currency) (decimal.Decimal, error) {
	if currency.MinorUnits <= 0 {
		return decimal.Zero, fmt.Errorf("invalid currency minor units: %d", currency.MinorUnits)
	}
    divisor := decimal.NewFromInt(10).Pow(decimal.NewFromInt(currency.MinorUnits))
	return decimal.NewFromInt(minorUnits).Div(divisor), nil
}

func (uc *walletUsecase) checkBalance(wallet *models.Wallet, amount int64, opType models.OperationType) (int64, error) {
    if opType == models.OperationWithdraw && wallet.Balance < amount {
        uc.log.Warn("Insufficient funds",
            logger.Int64Field("balance", wallet.Balance),
            logger.Int64Field("requested", amount))
        return 0, ErrInsufficientFunds
    }
    return amount, nil
}

