package reward

import (
	"github.com/fabricioricard/nkn-defi/internal/ports/repo"
	"github.com/fabricioricard/nkn-defi/internal/ports/eventbus"
)

type Engine struct{}

func NewEngine(swapRepo repo.SwapRepository, poolRepo repo.PoolRepository, bus eventbus.Publisher) *Engine {
	return &Engine{}
}