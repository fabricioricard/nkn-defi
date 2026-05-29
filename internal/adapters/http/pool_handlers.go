package http

import (
	"encoding/json"
	"math/big"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (rt *Router) createPool(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token0 string `json:"token0"`
		Token1 string `json:"token1"`
		FeeBps int    `json:"fee_bps"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pool, err := rt.poolUC.CreatePool(r.Context(), req.Token0, req.Token1, req.FeeBps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pool)
}

func (rt *Router) listPools(w http.ResponseWriter, r *http.Request) {
	pools, err := rt.poolUC.ListPools(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pools)
}

func (rt *Router) getPool(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	pool, err := rt.poolUC.GetPool(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pool)
}

func (rt *Router) addLiquidity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Amount0 string `json:"amount0"`
		Amount1 string `json:"amount1"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	amt0, ok0 := new(big.Int).SetString(req.Amount0, 10)
	amt1, ok1 := new(big.Int).SetString(req.Amount1, 10)
	if !ok0 || !ok1 {
		http.Error(w, "invalid amounts", http.StatusBadRequest)
		return
	}
	pool, lp, err := rt.poolUC.AddLiquidity(r.Context(), id, "anonymous", amt0, amt1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := map[string]interface{}{
		"pool":      pool,
		"lp_tokens": lp.String(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (rt *Router) swap(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		TokenIn  string `json:"token_in"`
		AmountIn string `json:"amount_in"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	amtIn, ok := new(big.Int).SetString(req.AmountIn, 10)
	if !ok {
		http.Error(w, "invalid amount_in", http.StatusBadRequest)
		return
	}
	amountOut, fee, err := rt.poolUC.Swap(r.Context(), id, "anonymous", req.TokenIn, amtIn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := map[string]string{
		"amount_out": amountOut.String(),
		"fee":        fee.String(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}