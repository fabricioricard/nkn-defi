package http

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/fabricioricard/nkn-defi/internal/adapters/postgres"
	"github.com/fabricioricard/nkn-defi/internal/app"
)

type Router struct {
	poolUC     *app.PoolUsecase
	logg       *zap.Logger
	bridgeRepo *postgres.BridgeRepository
}

// NewRouter recebe o sistema de arquivos do frontend e o repositório da Bridge.
func NewRouter(
	poolUC *app.PoolUsecase,
	logg *zap.Logger,
	frontendFS fs.FS,
	bridgeRepo *postgres.BridgeRepository,
) chi.Router {
	r := chi.NewRouter()
	rt := &Router{
		poolUC:     poolUC,
		logg:       logg,
		bridgeRepo: bridgeRepo,
	}

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/pools", rt.createPool)
		r.Get("/pools", rt.listPools)
		r.Get("/pools/{id}", rt.getPool)
		r.Post("/pools/{id}/liquidity/add", rt.addLiquidity)
		r.Post("/pools/{id}/swap", rt.swap)

		// Bridge (implementação real)
		r.Post("/bridge/deposit", rt.createBridgeDeposit)
		r.Get("/bridge/transactions", rt.getBridgeTransactions)
	})

	// Serve o frontend React (Single Page Application)
	r.Handle("/*", spaHandler(frontendFS))
	return r
}

// spaHandler serve o index.html para qualquer rota que não corresponda a um arquivo estático.
func spaHandler(fsys fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		f, err := fsys.Open(path)
		if err != nil {
			serveIndex(fsys, w)
			return
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil || stat.IsDir() {
			serveIndex(fsys, w)
			return
		}

		http.FileServer(http.FS(fsys)).ServeHTTP(w, r)
	}
}

// serveIndex lê o index.html do sistema de arquivos e o escreve na resposta.
func serveIndex(fsys fs.FS, w http.ResponseWriter) {
	data, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

// createBridgeDeposit cria um novo pedido de depósito.
func (rt *Router) createBridgeDeposit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EthAddress string `json:"eth_address"`
		Amount     string `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.EthAddress == "" || req.Amount == "" {
		http.Error(w, "eth_address and amount are required", http.StatusBadRequest)
		return
	}

	depositID := uuid.New().String()
	nknAddress := os.Getenv("BRIDGE_NKN_ADDRESS") // endereço fixo da carteira custodiada
	if nknAddress == "" {
		http.Error(w, "bridge NKN address not configured", http.StatusInternalServerError)
		return
	}

	dep := &postgres.BridgeDeposit{
		ID:                depositID,
		EthAddress:        req.EthAddress,
		Amount:            req.Amount,
		NknDepositAddress: nknAddress,
		Status:            "pending",
		Memo:              depositID, // o próprio ID serve como memo
	}

	if err := rt.bridgeRepo.InsertDeposit(r.Context(), dep); err != nil {
		rt.logg.Error("failed to insert deposit", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := map[string]string{
		"deposit_id":      depositID,
		"deposit_address": nknAddress,
		"memo":            depositID,
		"expires_at":      time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// getBridgeTransactions retorna as últimas transações do usuário.
// Espera um query parameter "eth_address".
func (rt *Router) getBridgeTransactions(w http.ResponseWriter, r *http.Request) {
	ethAddress := r.URL.Query().Get("eth_address")
	if ethAddress == "" {
		http.Error(w, "eth_address query parameter is required", http.StatusBadRequest)
		return
	}

	// Busca depósitos e retiradas reais
	deposits, err := rt.bridgeRepo.ListDepositsByEthAddress(r.Context(), ethAddress)
	if err != nil {
		rt.logg.Error("failed to list deposits", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	withdrawals, err := rt.bridgeRepo.ListWithdrawalsByFromAddress(r.Context(), ethAddress)
	if err != nil {
		rt.logg.Error("failed to list withdrawals", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Formato unificado esperado pelo frontend
	type bridgeTx struct {
		ID        string `json:"id"`
		Type      string `json:"type"`
		Amount    string `json:"amount"`
		Status    string `json:"status"`
		Timestamp string `json:"timestamp"`
	}
	var txs []bridgeTx

	for _, d := range deposits {
		txs = append(txs, bridgeTx{
			ID:        d.ID,
			Type:      "deposit",
			Amount:    d.Amount,
			Status:    d.Status,
			Timestamp: d.CreatedAt.Format(time.RFC3339),
		})
	}
	for _, w := range withdrawals {
		txs = append(txs, bridgeTx{
			ID:        w.ID,
			Type:      "withdrawal",
			Amount:    w.Amount,
			Status:    w.Status,
			Timestamp: w.CreatedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(txs)
}