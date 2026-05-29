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
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NKN DeFi · InfraFi</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root {
            --bg: #0b0b1a;
            --surface: #13132b;
            --primary: #7c3aed;
            --secondary: #a78bfa;
            --accent: #f59e0b;
            --text: #e2e8f0;
            --text-muted: #94a3b8;
        }
        body {
            background: var(--bg);
            color: var(--text);
            font-family: 'Inter', sans-serif;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
        }
        .navbar {
            background: rgba(19, 19, 43, 0.8);
            backdrop-filter: blur(10px);
            border-bottom: 1px solid rgba(124, 58, 237, 0.2);
        }
        .navbar-brand {
            font-weight: 700;
            color: var(--primary) !important;
            letter-spacing: -0.5px;
        }
        .container-main {
            flex: 1;
            padding-top: 30px;
            padding-bottom: 40px;
        }
        .card-pool {
            background: var(--surface);
            border: 1px solid rgba(124, 58, 237, 0.15);
            border-radius: 16px;
            transition: all 0.3s ease;
            margin-bottom: 20px;
        }
        .card-pool:hover {
            border-color: var(--primary);
            box-shadow: 0 10px 25px -5px rgba(124, 58, 237, 0.3);
            transform: translateY(-2px);
        }
        .btn-primary-custom {
            background: var(--primary);
            border: none;
            color: white;
            font-weight: 600;
            border-radius: 10px;
            padding: 8px 20px;
            transition: all 0.2s;
        }
        .btn-primary-custom:hover {
            background: #6d28d9;
            box-shadow: 0 4px 12px rgba(124, 58, 237, 0.5);
        }
        .btn-outline-custom {
            border: 1px solid var(--primary);
            color: var(--primary);
            background: transparent;
            border-radius: 10px;
            padding: 8px 20px;
            font-weight: 600;
            transition: all 0.2s;
        }
        .btn-outline-custom:hover {
            background: var(--primary);
            color: white;
        }
        .input-custom {
            background: rgba(255,255,255,0.05);
            border: 1px solid rgba(124, 58, 237, 0.3);
            color: var(--text);
            border-radius: 10px;
            padding: 10px 15px;
            transition: 0.2s;
        }
        .input-custom:focus {
            border-color: var(--primary);
            box-shadow: 0 0 0 0.2rem rgba(124, 58, 237, 0.25);
            background: rgba(255,255,255,0.08);
        }
        .select-custom {
            background: rgba(255,255,255,0.05);
            border: 1px solid rgba(124, 58, 237, 0.3);
            color: var(--text);
            border-radius: 10px;
            padding: 10px 15px;
            transition: 0.2s;
        }
        .section-title {
            font-weight: 700;
            font-size: 1.5rem;
            margin-bottom: 1.5rem;
            color: var(--secondary);
            border-left: 4px solid var(--primary);
            padding-left: 15px;
        }
        .footer {
            background: rgba(19, 19, 43, 0.6);
            border-top: 1px solid rgba(124, 58, 237, 0.2);
            color: var(--text-muted);
            padding: 20px 0;
            text-align: center;
        }
        .token-symbol {
            font-weight: 700;
            background: rgba(124, 58, 237, 0.15);
            padding: 2px 10px;
            border-radius: 20px;
            font-size: 0.9rem;
        }
        .reserve-text {
            font-family: monospace;
            color: var(--accent);
        }
        .lp-badge {
            background: rgba(245, 158, 11, 0.15);
            color: var(--accent);
            border-radius: 20px;
            padding: 2px 12px;
            font-size: 0.85rem;
            font-weight: 600;
        }
    </style>
</head>
<body>
    <nav class="navbar navbar-expand-lg navbar-dark sticky-top">
        <div class="container">
            <a class="navbar-brand" href="#"><i class="fas fa-cubes me-2"></i>NKN DeFi · InfraFi</a>
            <div class="d-flex gap-2">
                <button class="btn btn-outline-custom btn-sm"><i class="fas fa-chart-line me-1"></i>Dashboard</button>
                <button class="btn btn-outline-custom btn-sm"><i class="fas fa-book-open me-1"></i>Docs</button>
            </div>
        </div>
    </nav>

    <div class="container container-main">
        <div class="row mb-5">
            <div class="col-lg-5 mx-auto">
                <h5 class="section-title"><i class="fas fa-plus-circle me-2"></i>Create New Pool</h5>
                <div class="card-pool p-4">
                    <div class="mb-3">
                        <label class="form-label fw-semibold">Token Pair</label>
                        <div class="input-group">
                            <input type="text" id="token0" class="form-control input-custom" placeholder="e.g., NKN">
                            <span class="input-group-text bg-transparent border-0 text-light">/</span>
                            <input type="text" id="token1" class="form-control input-custom" placeholder="e.g., USDT">
                        </div>
                    </div>
                    <div class="mb-3">
                        <label class="form-label fw-semibold">Fee (bps)</label>
                        <input type="number" id="feeBps" class="form-control input-custom" value="30" min="1" max="1000">
                        <small class="text-muted">Standard: 30 bps (0.3%)</small>
                    </div>
                    <button class="btn btn-primary-custom w-100" onclick="createPool()">
                        <i class="fas fa-plus me-2"></i>Create Pool
                    </button>
                </div>
            </div>
        </div>

        <h5 class="section-title"><i class="fas fa-droplet me-2"></i>Liquidity Pools</h5>
        <div id="pools" class="row"></div>
    </div>

    <footer class="footer">
        <div class="container">
            <span>© 2026 NKN DeFi · Powered by InfraFi</span>
        </div>
    </footer>

    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
    <script>
        const API = '/api/v1';

        async function fetchJSON(url, options) {
            const res = await fetch(url, options);
            return res.json();
        }

        async function loadPools() {
            try {
                const pools = await fetchJSON(API + '/pools');
                const container = document.getElementById('pools');
                if (!pools || pools.length === 0) {
                    container.innerHTML = ` + "`" + `
                        <div class="col-12 text-center py-5">
                            <i class="fas fa-database fa-3x text-muted mb-3"></i>
                            <p class="text-muted">No pools created yet. Be the first to add liquidity!</p>
                        </div>` + "`" + `;
                    return;
                }
                container.innerHTML = pools.map(p => ` + "`" + `
                    <div class="col-md-6 col-lg-4 mb-4">
                        <div class="card-pool p-3 h-100">
                            <div class="d-flex justify-content-between align-items-center mb-3">
                                <span class="fw-bold fs-5">${escapeHTML(p.token0)}<span class="text-muted">/</span>${escapeHTML(p.token1)}</span>
                                <span class="lp-badge"><i class="fas fa-coins me-1"></i>${p.total_lp_tokens}</span>
                            </div>
                            <div class="mb-3">
                                <div class="d-flex justify-content-between small mb-1">
                                    <span>${escapeHTML(p.token0)} Reserve</span>
                                    <span class="reserve-text">${p.reserve0}</span>
                                </div>
                                <div class="d-flex justify-content-between small mb-1">
                                    <span>${escapeHTML(p.token1)} Reserve</span>
                                    <span class="reserve-text">${p.reserve1}</span>
                                </div>
                                <div class="d-flex justify-content-between small">
                                    <span>Fee</span>
                                    <span>${p.fee_bps} bps</span>
                                </div>
                            </div>
                            <div class="mt-3">
                                <h6 class="text-muted small mb-2">Add Liquidity</h6>
                                <div class="input-group input-group-sm mb-2">
                                    <input type="text" id="amount0-${p.id}" class="form-control input-custom" placeholder="${escapeHTML(p.token0)} amount">
                                    <input type="text" id="amount1-${p.id}" class="form-control input-custom" placeholder="${escapeHTML(p.token1)} amount">
                                    <button class="btn btn-primary-custom" onclick="addLiquidity('${p.id}')"><i class="fas fa-plus"></i></button>
                                </div>
                                <h6 class="text-muted small mb-2 mt-3">Swap</h6>
                                <div class="input-group input-group-sm">
                                    <input type="text" id="swapIn-${p.id}" class="form-control input-custom" placeholder="Amount">
                                    <select id="tokenIn-${p.id}" class="form-select select-custom" style="max-width:100px;">
                                        <option>${escapeHTML(p.token0)}</option>
                                        <option>${escapeHTML(p.token1)}</option>
                                    </select>
                                    <button class="btn btn-outline-custom" onclick="doSwap('${p.id}')"><i class="fas fa-exchange-alt"></i></button>
                                </div>
                            </div>
                        </div>
                    </div>
                ` + "`" + `).join('');
            } catch (err) {
                console.error(err);
            }
        }

        function escapeHTML(str) {
            return String(str).replace(/[&<>"']/g, function(m) {
                return { '&':'&amp;', '<':'&lt;', '>':'&gt;', '"':'&quot;', "'":'&#39;' }[m];
            });
        }

        async function createPool() {
            const t0 = document.getElementById('token0').value.trim();
            const t1 = document.getElementById('token1').value.trim();
            const fee = parseInt(document.getElementById('feeBps').value);
            if (!t0 || !t1) return alert('Please fill both token symbols');
            if (isNaN(fee) || fee <= 0) return alert('Invalid fee');
            await fetchJSON(API + '/pools', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ token0: t0, token1: t1, fee_bps: fee })
            });
            loadPools();
        }

        async function addLiquidity(poolId) {
            const amount0 = document.getElementById('amount0-' + poolId)?.value;
            const amount1 = document.getElementById('amount1-' + poolId)?.value;
            if (!amount0 || !amount1) return alert('Enter both amounts');
            await fetchJSON(API + '/pools/' + poolId + '/liquidity/add', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ amount0, amount1 })
            });
            loadPools();
        }

        async function doSwap(poolId) {
            const amountIn = document.getElementById('swapIn-' + poolId)?.value;
            const tokenIn = document.getElementById('tokenIn-' + poolId)?.value;
            if (!amountIn) return alert('Enter amount');
            const res = await fetchJSON(API + '/pools/' + poolId + '/swap', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ token_in: tokenIn, amount_in: amountIn })
            });
            alert(` + "`" + `Swap successful!\nOutput: ${res.amount_out}\nFee: ${res.fee}` + "`" + `);
            loadPools();
        }

        loadPools();
    </script>
</body>
</html>`