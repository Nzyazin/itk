package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

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
	var operation models.WalletOperation
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&operation); err != nil {
		h.log.Warn("Failed to decode request body", logger.ErrorField("error", err))
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if operation.WalletID == uuid.Nil {
		h.log.Warn("Wallet ID is required")
		respondWithError(w, http.StatusBadRequest, "Wallet ID is required")
		return
	}

	if operation.OperationType != models.OperationDeposit && operation.OperationType != models.OperationWithdraw {
		h.log.Warn("Invalid operation type", logger.StringField("operation_type", string(operation.OperationType)))
		respondWithError(w, http.StatusBadRequest, "Invalid operation type. Must be DEPOSIT or WITHDRAW")
		return
	}

	amountStr := strings.ReplaceAll(operation.Amount, " ", "")
	amountStr = strings.ReplaceAll(amountStr, ",", ".")

	if !(amountRegexp.MatchString(amountStr)) {
		h.log.Warn("Invalid amount format", logger.StringField("amount", amountStr))
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid amount format: %s", amountStr))
		return
	}

	amountDec, err := decimal.NewFromString(amountStr)
	if err != nil {
		h.log.Warn("Failed to parse amount", logger.StringField("amount", amountStr), logger.ErrorField("error", err))
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("could not parse amount: %s", err))
		return
	}

	if amountDec.LessThanOrEqual(decimal.NewFromInt(0)) {
		h.log.Warn("Non-positive amount in request", logger.StringField("amount", amountDec.String()))
		respondWithError(w, http.StatusBadRequest, "Amount must be positive")
		return
	}
	operation.DecimalAmount = amountDec
	newBalance, err := h.usecase.OperateWallet(r.Context(), operation)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrWalletNotFound) || strings.Contains(err.Error(), "not found"):
			h.log.Warn("Wallet not found", logger.StringField("wallet_id", operation.WalletID.String()))
			respondWithError(w, http.StatusNotFound, "Wallet not found")
		case errors.Is(err, usecase.ErrInvalidAmount):
			h.log.Warn("Invalid amount", logger.StringField("amount", operation.DecimalAmount.String()))
			respondWithError(w, http.StatusBadRequest, "Invalid amount")
		case errors.Is(err, usecase.ErrInsufficientFunds) || strings.Contains(err.Error(), "insufficient funds"):
			h.log.Warn("Insufficient funds", logger.StringField("wallet_id", operation.WalletID.String()), logger.StringField("amount", operation.DecimalAmount.String()))
			respondWithError(w, http.StatusBadRequest, "Insufficient funds")
		default:
			h.log.Error("Failed to process operation", logger.StringField("wallet_id", operation.WalletID.String()), logger.StringField("amount", operation.DecimalAmount.String()), logger.ErrorField("error", err))
			respondWithError(w, http.StatusInternalServerError, "Failed to process operation")
		}
		return
	}
	balanceStr := newBalance.StringFixedBank(2)

	h.log.Info("Wallet operation successful",
		logger.StringField("wallet_id", operation.WalletID.String()),
		logger.StringField("operation_type", string(operation.OperationType)),
		logger.StringField("amount", amountDec.String()),
		logger.StringField("new_balance", balanceStr),
	)

	respondWithJSON(w, http.StatusOK, OperationResponse{
		Error: "",
		Balance: balanceStr,
		WalletID: operation.WalletID,
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
