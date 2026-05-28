package sync

import (
	"context"
	"database/sql"
	"github.com/fabricioricard/nkn-defi/internal/ports/cache"
	"github.com/fabricioricard/nkn-defi/internal/ports/eventbus"
	"github.com/fabricioricard/nkn-defi/internal/adapters/nknclient"
	"github.com/fabricioricard/nkn-defi/internal/app"
)

type BlockSyncer struct{}

func NewBlockSyncer(nknClient *nknclient.Client, bus eventbus.Publisher, db *sql.DB, cache cache.Cache, poolUC *app.PoolUsecase) *BlockSyncer {
	return &BlockSyncer{}
}

func (bs *BlockSyncer) Start(ctx context.Context) {}
func (bs *BlockSyncer) Stop()                     {}