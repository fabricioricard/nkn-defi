package pricing

import (
	"github.com/fabricioricard/nkn-defi/internal/ports/cache"
	"github.com/fabricioricard/nkn-defi/internal/ports/repo"
	"github.com/fabricioricard/nkn-defi/internal/adapters/nknclient"
)

type Oracle struct{}

func NewOracle(cache cache.Cache, poolRepo repo.PoolRepository, nknClient *nknclient.Client) *Oracle {
	return &Oracle{}
}