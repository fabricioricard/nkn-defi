package http

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/fabricioricard/nkn-defi/internal/app"
)

type Router struct {
	poolUC *app.PoolUsecase
	logg   *zap.Logger
}

func NewRouter(poolUC *app.PoolUsecase, logg *zap.Logger) chi.Router {
	r := chi.NewRouter()
	rt := &Router{poolUC: poolUC, logg: logg}

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/pools", rt.createPool)
		r.Get("/pools", rt.listPools)
		r.Get("/pools/{id}", rt.getPool)
		r.Post("/pools/{id}/liquidity/add", rt.addLiquidity)
		r.Post("/pools/{id}/swap", rt.swap)
	})

	r.Get("/", rt.serveFrontend)
	return r
}

func (rt *Router) serveFrontend(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, frontendHTML)
}

const frontendHTML = `<!DOCTYPE html>
<html>
<head>
    <title>NKN DeFi - InfraFi</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial; max-width: 800px; margin: auto; padding: 20px; }
        .pool { border: 1px solid #ccc; padding: 10px; margin: 10px 0; }
        .actions { margin-top: 10px; }
        input { margin: 5px; padding: 5px; width: 150px; }
        button { padding: 5px 10px; }
    </style>
</head>
<body>
    <h1>NKN DeFi (InfraFi)</h1>
    <div>
        <h2>Create Pool</h2>
        <input type="text" id="token0" placeholder="Token0 (e.g., NKN)">
        <input type="text" id="token1" placeholder="Token1 (e.g., USDT)">
        <input type="number" id="feeBps" placeholder="Fee (bps)" value="30">
        <button onclick="createPool()">Create</button>
    </div>
    <h2>Pools</h2>
    <div id="pools"></div>

    <script>
        const API = '/api/v1';

        async function fetchJSON(url, options) {
            const res = await fetch(url, options);
            return res.json();
        }

        async function loadPools() {
            const pools = await fetchJSON(API + '/pools');
            const div = document.getElementById('pools');
            div.innerHTML = pools.map(p => ` + "`" + `
                <div class="pool">
                    <strong>${p.token0}/${p.token1}</strong> (ID: ${p.id})<br>
                    Reserve0: ${p.reserve0}, Reserve1: ${p.reserve1}<br>
                    Fee: ${p.fee_bps} bps, LP Tokens: ${p.total_lp_tokens}
                    <div class="actions">
                        <input type="text" id="amount0-${p.id}" placeholder="Amount ${p.token0}">
                        <input type="text" id="amount1-${p.id}" placeholder="Amount ${p.token1}">
                        <button onclick="addLiquidity('${p.id}')">Add Liquidity</button>
                        <input type="text" id="swapIn-${p.id}" placeholder="Amount In">
                        <select id="tokenIn-${p.id}"><option>${p.token0}</option><option>${p.token1}</option></select>
                        <button onclick="doSwap('${p.id}')">Swap</button>
                    </div>
                </div>` + "`" + `).join('');
        }

        async function createPool() {
            const t0 = document.getElementById('token0').value;
            const t1 = document.getElementById('token1').value;
            const fee = document.getElementById('feeBps').value;
            if (!t0 || !t1) return alert('Fill tokens');
            await fetchJSON(API + '/pools', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ token0: t0, token1: t1, fee_bps: parseInt(fee) })
            });
            loadPools();
        }

        async function addLiquidity(poolId) {
            const amount0 = document.getElementById('amount0-' + poolId).value;
            const amount1 = document.getElementById('amount1-' + poolId).value;
            if (!amount0 || !amount1) return alert('Enter amounts');
            await fetchJSON(API + '/pools/' + poolId + '/liquidity/add', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ amount0, amount1 })
            });
            loadPools();
        }

        async function doSwap(poolId) {
            const amountIn = document.getElementById('swapIn-' + poolId).value;
            const tokenIn = document.getElementById('tokenIn-' + poolId).value;
            if (!amountIn) return alert('Enter amount');
            const res = await fetchJSON(API + '/pools/' + poolId + '/swap', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ token_in: tokenIn, amount_in: amountIn })
            });
            alert('Output: ' + res.amount_out + ', Fee: ' + res.fee);
            loadPools();
        }

        loadPools();
    </script>
</body>
</html>`