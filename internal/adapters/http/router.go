package http

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    "go.uber.org/zap"

    "github.com/fabricioricard/nkn-defi/internal/app"
    "github.com/fabricioricard/nkn-defi/internal/ports/cache"
)

func NewRouter(
    poolUC *app.PoolUsecase,
    swapUC *app.SwapUsecase,
    stakeUC *app.StakingUsecase,
    nodeUC *app.NodeUsecase,
    logg *zap.Logger,
    cache cache.Cache,
) chi.Router {
    r := chi.NewRouter()

    // Health check
    r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

    // Futuras rotas virão aqui...

    return r
}