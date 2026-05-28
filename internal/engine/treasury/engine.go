package treasury

import (
	"github.com/fabricioricard/nkn-defi/internal/ports/repo"
	"github.com/fabricioricard/nkn-defi/internal/ports/eventbus"
)

type Engine struct{}

func NewEngine(treasuryRepo repo.TreasuryRepository, bus eventbus.Publisher) *Engine {
	return &Engine{}
}