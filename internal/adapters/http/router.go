package http

import (
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/fabricioricard/nkn-defi/internal/adapters/ethereum"
	"github.com/fabricioricard/nkn-defi/internal/adapters/nknclient"
	"github.com/fabricioricard/nkn-defi/internal/adapters/postgres"
	"github.com/fabricioricard/nkn-defi/internal/app"
)

type Router struct {
	poolUC     *app.PoolUsecase
	logg       *zap.Logger
	bridgeRepo *postgres.BridgeRepository
	ethClient  *ethereum.Client
}

// NewRouter recebe o sistema de arquivos do frontend, o repositório da Bridge e o cliente Ethereum.
func NewRouter(
	poolUC *app.PoolUsecase,
	logg *zap.Logger,
	frontendFS fs.FS,
	bridgeRepo *postgres.BridgeRepository,
	ethClient *ethereum.Client,
) chi.Router {
	r := chi.NewRouter()
	rt := &Router{
		poolUC:     poolUC,
		logg:       logg,
		bridgeRepo: bridgeRepo,
		ethClient:  ethClient,
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

		// Bridge
		r.Post("/bridge/deposit", rt.createBridgeDeposit)
		r.Get("/bridge/transactions", rt.getBridgeTransactions)
		r.Delete("/bridge/deposit/{id}", rt.cancelBridgeDeposit)
	})

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

// createBridgeDeposit cria um novo pedido de depósito com endereço NKN único.
// Se o hash da transação NKN for fornecido, o mint é feito imediatamente.
func (rt *Router) createBridgeDeposit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EthAddress string `json:"eth_address"`
		Amount     string `json:"amount"`
		TxHash     string `json:"tx_hash,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.EthAddress == "" || req.Amount == "" {
		http.Error(w, "eth_address and amount are required", http.StatusBadRequest)
		return
	}

	// Obtém a chave privada da carteira NKN principal
	privateKey := os.Getenv("BRIDGE_NKN_PRIVATE_KEY")
	if privateKey == "" {
		rt.logg.Error("BRIDGE_NKN_PRIVATE_KEY not set")
		http.Error(w, "bridge not configured", http.StatusInternalServerError)
		return
	}

	wallet, err := nknclient.NewWallet(privateKey)
	if err != nil {
		rt.logg.Error("failed to create NKN wallet", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Gera um endereço único derivado do timestamp (índice)
	index := uint32(time.Now().UnixMilli())
	derivedAddr, err := wallet.DeriveAddress(index)
	if err != nil {
		rt.logg.Error("failed to derive NKN address", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	depositID := uuid.New().String()
	dep := &postgres.BridgeDeposit{
		ID:                depositID,
		EthAddress:        req.EthAddress,
		Amount:            req.Amount,
		NknDepositAddress: derivedAddr,
		Status:            "pending",
		Memo:              depositID,
		NknDerivedIndex:   int64(index),
	}

	if err := rt.bridgeRepo.InsertDeposit(r.Context(), dep); err != nil {
		rt.logg.Error("failed to insert deposit", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Se o hash da transação NKN for fornecido, realiza o mint imediatamente
	if req.TxHash != "" {
		// Converte o hash para bytes32
		hashBytes, err := hexStringToBytes32(req.TxHash)
		if err != nil {
			rt.logg.Error("invalid tx hash", zap.Error(err))
			http.Error(w, "invalid tx_hash", http.StatusBadRequest)
			return
		}

		// Converte a quantidade para *big.Int
		amountWei, ok := new(big.Int).SetString(req.Amount, 10)
		if !ok {
			http.Error(w, "invalid amount", http.StatusBadRequest)
			return
		}

		// Chama o mint no contrato wNKN
		mintTxHash, err := rt.ethClient.MintWNKN(
			common.HexToAddress(req.EthAddress),
			amountWei,
			hashBytes,
		)
		if err != nil {
			rt.logg.Error("failed to mint wNKN", zap.Error(err))
			http.Error(w, "mint failed", http.StatusInternalServerError)
			return
		}

		// Atualiza o depósito para completed
		if err := rt.bridgeRepo.UpdateDepositAfterMint(
			r.Context(),
			dep.ID,
			req.TxHash,
			mintTxHash,
			"completed",
		); err != nil {
			rt.logg.Error("failed to update deposit after mint", zap.Error(err))
		}

		// Retorna resposta com sucesso
		resp := map[string]string{
			"deposit_id":      depositID,
			"deposit_address": derivedAddr,
			"mint_tx_hash":    mintTxHash,
			"status":          "completed",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Fluxo normal (sem tx_hash): apenas gera o endereço e fica pendente
	resp := map[string]string{
		"deposit_id":      depositID,
		"deposit_address": derivedAddr,
		"memo":            "",
		"expires_at":      time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// cancelBridgeDeposit cancela um depósito pendente.
func (rt *Router) cancelBridgeDeposit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := rt.bridgeRepo.DeleteDeposit(r.Context(), id); err != nil {
		rt.logg.Error("failed to delete deposit", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
}

// getBridgeTransactions retorna as últimas transações do usuário.
func (rt *Router) getBridgeTransactions(w http.ResponseWriter, r *http.Request) {
	ethAddress := r.URL.Query().Get("eth_address")
	if ethAddress == "" {
		http.Error(w, "eth_address query parameter is required", http.StatusBadRequest)
		return
	}

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

// hexStringToBytes32 converte uma string hexadecimal (com ou sem 0x) para [32]byte.
func hexStringToBytes32(hexStr string) ([32]byte, error) {
	var b [32]byte
	clean := hexStr
	if len(clean) >= 2 && clean[:2] == "0x" {
		clean = clean[2:]
	}
	bytes, err := hex.DecodeString(clean)
	if err != nil {
		return b, err
	}
	copy(b[:], bytes)
	return b, nil
}