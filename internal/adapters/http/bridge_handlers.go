package http

import (
	"encoding/json"
	"net/http"
	"time"
	"github.com/google/uuid"
)

type depositRequest struct {
	EthAddress string `json:"eth_address"`
	Amount     string `json:"amount"`
}

type depositResponse struct {
	DepositID       string `json:"deposit_id"`
	DepositAddress  string `json:"deposit_address"`
	ExpiresAt       string `json:"expires_at"`
}

func (rt *Router) createBridgeDeposit(w http.ResponseWriter, r *http.Request) {
	var req depositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	// Valida endereço Ethereum
	// Gera deposit_id e armazena no banco/Redis com mapeamento ethAddress
	depositID := uuid.New().String()
	expiration := time.Now().Add(1 * time.Hour)
	// AQUI: inserir no BD (tabela bridge_deposits) com status "pending"

	resp := depositResponse{
		DepositID:      depositID,
		DepositAddress: getBridgeNKNAddress(), // endereço fixo ou derivado
		ExpiresAt:      expiration.Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (rt *Router) getBridgeTransactions(w http.ResponseWriter, r *http.Request) {
	// consultar tabela bridge_transactions e retornar
}