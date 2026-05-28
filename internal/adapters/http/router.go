package http

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"github.com/fabricioricard/nkn-defi/internal/ports/cache"
	"github.com/fabricioricard/nkn-defi/internal/app"
)

func NewRouter(poolUC *app.PoolUsecase, swapUC *app.SwapUsecase, stakeUC *app.StakingUsecase, nodeUC *app.NodeUsecase, logg *zap.Logger, cache cache.Cache) chi.Router {
	r := chi.NewRouter()
	// rotas serão adicionadas depois
	return r
}