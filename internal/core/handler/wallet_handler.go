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
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

// WalletHandler обрабатывает HTTP-запросы для операций с кошельками
type WalletHandler struct {
	usecase usecase.WalletUsecase
}

type OperationResponse struct {
	Error string `json:"error,omitempty"`
	Balance string `json:"balance"`
	WalletID uuid.UUID `json:"wallet_id"`
}

func NewWalletHandler(usecase usecase.WalletUsecase) *WalletHandler {
	return &WalletHandler{usecase: usecase}
}

func (h *WalletHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/wallet", h.ProcessWalletOperation).Methods("POST")
	// router.HandleFunc("/api/v1/wallets/{wallet_id}", h.GetWallet).Methods("GET")
}

func (h *WalletHandler) ProcessWalletOperation(w http.ResponseWriter, r *http.Request) {
	var operation models.WalletOperation

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&operation); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if operation.WalletID == uuid.Nil {
		respondWithError(w, http.StatusBadRequest, "Wallet ID is required")
		return
	}

	if operation.OperationType != models.OperationDeposit && operation.OperationType != models.OperationWithdraw {
		respondWithError(w, http.StatusBadRequest, "Invalid operation type. Must be DEPOSIT or WITHDRAW")
		return
	}

	amountStr := strings.ReplaceAll(operation.Amount, " ", "")
	amountStr = strings.ReplaceAll(amountStr, ",", ".")

	acceptableNumbers := `^\s*\d{1,9}([.,]\d{1,2})?\s*$`
	if !(regexp.MustCompile(acceptableNumbers).MatchString(amountStr)) {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid amount format: %s", amountStr))
		return
	}

	amountDec, err := decimal.NewFromString(amountStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("could not parse amount: %s", err))
		return
	}

	if amountDec.LessThanOrEqual(decimal.NewFromInt(0)) {
		respondWithError(w, http.StatusBadRequest, "Amount must be positive")
		return
	}
	operation.DecimalAmount = amountDec
	newBalance, err := h.usecase.OperateWallet(r.Context(), operation)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrWalletNotFound) || strings.Contains(err.Error(), "not found"):
			respondWithError(w, http.StatusNotFound, "Wallet not found")
		case errors.Is(err, usecase.ErrInvalidAmount):
			respondWithError(w, http.StatusBadRequest, "Invalid amount")
		case errors.Is(err, usecase.ErrInsufficientFunds) || strings.Contains(err.Error(), "insufficient funds"):
			respondWithError(w, http.StatusBadRequest, "Insufficient funds")
		default:
			respondWithError(w, http.StatusInternalServerError, "Failed to process operation")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, OperationResponse{
		Error: "",
		Balance: fmt.Sprintf("%d", newBalance),
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
