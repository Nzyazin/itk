package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Nzyazin/itk/internal/core/repository"
	"github.com/Nzyazin/itk/internal/core/logger"
	"github.com/Nzyazin/itk/internal/core/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
    transactionStatusCompleted = "COMPLETED"
    transactionStatusFailed    = "FAILED"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidAmount     = errors.New("amount must be positive")
	ErrInvalidOperationType = errors.New("invalid operation type")
)

type postgresWalletRepo struct {
	db *sqlx.DB
	log logger.Logger
}

func NewPostgresWalletRepo(db *sqlx.DB, log logger.Logger) repository.WalletRepository {
	return &postgresWalletRepo{
		db: db,
		log: log,
	}
}

func (r *postgresWalletRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	var wallet models.Wallet
	query := `SELECT id, balance, currency_code, created_at, updated_at FROM wallets WHERE id = $1`
	err := r.db.GetContext(ctx, &wallet, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: wallet with id %s not found", sql.ErrNoRows, id)
		}
		return nil, fmt.Errorf("error getting wallet: %w", err)
	}

	return &wallet, nil
}

func (r *postgresWalletRepo) GetCurrencyByCode(ctx context.Context, code string) (*models.Currency, error) {
	var currency models.Currency
	query := `SELECT code, name, minor_units FROM currencies WHERE code = $1`
	err := r.db.GetContext(ctx, &currency, query, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("currency with code %s not found", code)
		}
		return nil, fmt.Errorf("error getting currency: %w", err)
	}

	return &currency, nil
}

func (r *postgresWalletRepo) ExecuteTx(ctx context.Context, walletID uuid.UUID, amount int64, operationType models.OperationType) (int64, error) {
	var isCommitted bool
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
    if err != nil {
        r.log.Error("Error beginning transaction", 
            logger.ErrorField("error", err))
        return 0, fmt.Errorf("error beginning transaction: %w", err)
    }

    defer func() {
        if err != nil && !isCommitted {
            if rbErr := tx.Rollback(); rbErr != nil {
                r.log.Error("Transaction rollback failed", 
                    logger.ErrorField("error", rbErr))
                err = fmt.Errorf("%w (rollback failed: %v)", err, rbErr)
            } else {
                r.log.Warn("Transaction rolled back due to error", 
                    logger.ErrorField("error", err))
            }
        }
    }()

    newBalance, err := r.updateBalance(ctx, tx, walletID, amount, operationType)
    if err != nil {
        return 0, err
    }

    if err := r.createTransaction(ctx, tx, walletID, amount, operationType, transactionStatusCompleted); err != nil {
        return 0, err
    }

    if err = tx.Commit(); err != nil {
        r.log.Error("Error committing transaction", 
            logger.ErrorField("error", err))
        return 0, fmt.Errorf("commit failed: %w", err)
    }

    isCommitted = true
    return newBalance, nil
}

func (r *postgresWalletRepo) updateBalance(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, amount int64, operationType models.OperationType) (int64, error) {
    var delta int64
    switch operationType {
    case models.OperationDeposit:
        delta = amount
    case models.OperationWithdraw:
        delta = -amount
    default:
        return 0, ErrInvalidOperationType
    }

    var newBalance int64
    updateQuery := `
        UPDATE wallets
        SET balance = balance + $1
        WHERE id = $2
        RETURNING balance
    `
    err := tx.GetContext(ctx, &newBalance, updateQuery, delta, walletID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return 0, fmt.Errorf("wallet not found: %s", walletID)
        }
        return 0, fmt.Errorf("update balance: %w", err)
    }

    if newBalance < 0 {
        return 0, ErrInsufficientFunds
    }

    return newBalance, nil
}

func (r *postgresWalletRepo) createTransaction(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, amount int64, operationType models.OperationType, status string) error {
    transaction := &models.Transaction{
        ID:            uuid.New(),
        WalletID:      walletID,
        OperationType: operationType,
        Amount:        amount,
        Status:        status,
    }

    const query = `INSERT INTO transactions 
        (id, wallet_id, operation_type, amount, status) 
        VALUES ($1, $2, $3, $4, $5)`

    _, err := tx.ExecContext(ctx, query,
        transaction.ID,
        transaction.WalletID,
        transaction.OperationType,
        transaction.Amount,
        transaction.Status,
    )

    if err != nil {
        return fmt.Errorf("create transaction: %w", err)
    }

    return nil
}

// func (r *postgresWalletRepo) TransferFunds(ctx context.Context, from, to uuid.UUID, amount int64) error {
//     var lastErr error
    
//     for attempt := 1; attempt <= maxRetries; attempt++ {
//         err := r.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
//             // Логика транзакции
//             return nil
//         })
        
//         if err == nil {
//             return nil
//         }
        
//         if isRetryableError(err) {
//             sleep := time.Duration(attempt*attempt) * time.Second
//             time.Sleep(sleep)
//             lastErr = err
//             continue
//         }
        
//         return err
//     }
    
//     return fmt.Errorf("transaction failed after %d attempts: %w", maxRetries, lastErr)
// }

// func isRetryableError(err error) bool {
//     // 40001 - serialization failure (PostgreSQL)
//     // 40P01 - deadlock detected
//     if pgErr, ok := err.(*pgconn.PgError); ok {
//         return pgErr.Code == "40001" || pgErr.Code == "40P01"
//     }
//     return false
// }