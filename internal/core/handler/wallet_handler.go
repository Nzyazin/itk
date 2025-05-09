package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"context"

	"github.com/Nzyazin/itk/internal/core/models"
	"github.com/Nzyazin/itk/internal/core/usecase"
	"github.com/Nzyazin/itk/internal/core/logger"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

type WalletHandler struct {
	usecase usecase.WalletUsecase
	log logger.Logger
}

type OperationResponse struct {
	Error string `json:"error,omitempty"`
	Balance string `json:"balance"`
	WalletID uuid.UUID `json:"wallet_id"`
}

var amountRegexp = regexp.MustCompile(`^\s*\d{1,9}([.,]\d{1,2})?\s*$`)

func NewWalletHandler(usecase usecase.WalletUsecase, log logger.Logger) *WalletHandler {
	return &WalletHandler{usecase: usecase, log: log}
}

func (h *WalletHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/wallet", h.ProcessWalletOperation).Methods("POST")
	// router.HandleFunc("/api/v1/wallets/{wallet_id}", h.GetWallet).Methods("GET")
}

func (h *WalletHandler) ProcessWalletOperation(w http.ResponseWriter, r *http.Request) {
    operation, err := h.decodeRequest(w, r)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, err.Error())
        return
    }

    if validationErr := h.validateOperation(operation); validationErr != nil {
        h.log.Warn(validationErr.Message, validationErr.Fields...)
        respondWithError(w, http.StatusBadRequest, validationErr.Message)
        return
    }

    amountDec, err := h.parseAmount(operation.Amount)
    if err != nil {
        h.log.Warn("Invalid amount", logger.StringField("amount", operation.Amount), logger.ErrorField("error", err))
        respondWithError(w, http.StatusBadRequest, err.Error())
        return
    }
    operation.DecimalAmount = amountDec

    newBalance, err := h.executeWalletOperation(r.Context(), operation)
    if err != nil {
        h.handleOperationError(w, operation, err)
        return
    }

    h.logSuccess(operation, newBalance)
    h.sendSuccessResponse(w, operation, newBalance)
}

type ValidationError struct {
    Message string
    Fields  []logger.Field
}

func (h *WalletHandler) decodeRequest(w http.ResponseWriter, r *http.Request) (*models.WalletOperation, error) {
    var operation models.WalletOperation
    r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
    if err := json.NewDecoder(r.Body).Decode(&operation); err != nil {
        h.log.Warn("Failed to decode request body", logger.ErrorField("error", err))
        return nil, fmt.Errorf("invalid request payload")
    }
    defer r.Body.Close()
    return &operation, nil
}

// validateOperation выполняет базовую валидацию полей операции
func (h *WalletHandler) validateOperation(operation *models.WalletOperation) *ValidationError {
    if operation.WalletID == uuid.Nil {
        return &ValidationError{
            Message: "Wallet ID is required",
            Fields:  []logger.Field{logger.StringField("wallet_id", "")},
        }
    }

    operation.OperationType = models.OperationType(
        strings.ToUpper(string(operation.OperationType)),
    )

    switch operation.OperationType {
    case models.OperationDeposit, models.OperationWithdraw:
        return nil
    default:
        return &ValidationError{
            Message: "Invalid operation type",
            Fields: []logger.Field{
                logger.StringField("operation_type", string(operation.OperationType)),
            },
        }
    }
}

// parseAmount обрабатывает и валидирует сумму операции
func (h *WalletHandler) parseAmount(amountStr string) (decimal.Decimal, error) {
    cleaned := strings.ReplaceAll(strings.ReplaceAll(amountStr, " ", ""), ",", ".")
    
    if !amountRegexp.MatchString(cleaned) {
        return decimal.Zero, fmt.Errorf("invalid amount format: %s", cleaned)
    }

    amount, err := decimal.NewFromString(cleaned)
    if err != nil {
        return decimal.Zero, fmt.Errorf("could not parse amount: %v", err)
    }

    if amount.LessThanOrEqual(decimal.Zero) {
        return decimal.Zero, fmt.Errorf("amount must be positive")
    }

    return amount, nil
}

func (h *WalletHandler) executeWalletOperation(ctx context.Context, op *models.WalletOperation) (decimal.Decimal, error) {
    return h.usecase.OperateWallet(ctx, *op)
}

func (h *WalletHandler) handleOperationError(w http.ResponseWriter, op *models.WalletOperation, err error) {
    switch {
    case errors.Is(err, usecase.ErrWalletNotFound):
        h.log.Warn("Wallet not found", logger.StringField("wallet_id", op.WalletID.String()))
        respondWithError(w, http.StatusNotFound, "Wallet not found")
    case errors.Is(err, usecase.ErrInvalidAmount):
        h.log.Warn("Invalid amount", logger.StringField("amount", op.DecimalAmount.String()))
        respondWithError(w, http.StatusBadRequest, "Invalid amount")
    case errors.Is(err, usecase.ErrInsufficientFunds):
        h.log.Warn("Insufficient funds", 
            logger.StringField("wallet_id", op.WalletID.String()),
            logger.StringField("amount", op.DecimalAmount.String()),
        )
        respondWithJSON(w, http.StatusBadRequest, OperationResponse{
            Error:    "insufficient funds",
            Balance:  "",
            WalletID: op.WalletID,
        })
    default:
        h.log.Error("Failed to process operation", 
            logger.StringField("wallet_id", op.WalletID.String()),
            logger.StringField("amount", op.DecimalAmount.String()),
            logger.ErrorField("error", err),
        )
        respondWithError(w, http.StatusInternalServerError, "Failed to process operation")
    }
}

func (h *WalletHandler) logSuccess(op *models.WalletOperation, newBalance decimal.Decimal) {
    h.log.Info("Wallet operation successful",
        logger.StringField("wallet_id", op.WalletID.String()),
        logger.StringField("operation_type", string(op.OperationType)),
        logger.StringField("amount", op.DecimalAmount.String()),
        logger.StringField("new_balance", newBalance.StringFixedBank(2)),
    )
}

func (h *WalletHandler) sendSuccessResponse(w http.ResponseWriter, op *models.WalletOperation, balance decimal.Decimal) {
    respondWithJSON(w, http.StatusOK, OperationResponse{
        Error:    "",
        Balance:  balance.StringFixedBank(2),
        WalletID: op.WalletID,
    })
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, OperationResponse{Error: message})
}

func respondWithJSON(w http.ResponseWriter, code int, os OperationResponse) {
	response, err := json.Marshal(os)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal Server Error"}`)) // Fallback response
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
