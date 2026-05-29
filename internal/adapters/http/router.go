package http

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/fabricioricard/nkn-defi/internal/app"
)

type Router struct {
	poolUC *app.PoolUsecase
	logg   *zap.Logger
}

// NewRouter agora recebe o sistema de arquivos do frontend (React build)
func NewRouter(poolUC *app.PoolUsecase, logg *zap.Logger, frontendFS fs.FS) chi.Router {
	r := chi.NewRouter()
	rt := &Router{poolUC: poolUC, logg: logg}

	// Health check
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/pools", rt.createPool)
		r.Get("/pools", rt.listPools)
		r.Get("/pools/{id}", rt.getPool)
		r.Post("/pools/{id}/liquidity/add", rt.addLiquidity)
		r.Post("/pools/{id}/swap", rt.swap)

		// Bridge (stubs)
		r.Post("/bridge/deposit", rt.createBridgeDeposit)
		r.Get("/bridge/transactions", rt.getBridgeTransactions)
	})

	// Servir arquivos estáticos do frontend React
	// O FileServer vai procurar automaticamente o index.html na raiz
	r.Handle("/*", http.FileServer(http.FS(frontendFS)))

	return r
}

// --- Stubs da Bridge (mantidos até implementação real) ---

func (rt *Router) createBridgeDeposit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"deposit_id":      "demo-" + fmt.Sprint(time.Now().Unix()),
		"deposit_address": "NKN0000-EXAMPLE-ADDRESS",
		"expires_at":      time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	})
}

func (rt *Router) getBridgeTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]map[string]interface{}{
		{
			"id":        "1",
			"type":      "deposit",
			"amount":    "1000",
			"status":    "completed",
			"timestamp": time.Now().Format(time.RFC3339),
		},
		{
			"id":        "2",
			"type":      "withdrawal",
			"amount":    "500",
			"status":    "pending",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	})
}