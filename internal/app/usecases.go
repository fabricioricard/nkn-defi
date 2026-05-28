package app

import (
    "github.com/fabricioricard/nkn-defi/internal/ports/repo"
    "github.com/fabricioricard/nkn-defi/internal/ports/cache"
    "github.com/fabricioricard/nkn-defi/internal/ports/eventbus"
    "github.com/fabricioricard/nkn-defi/internal/engine/amm"
    "github.com/fabricioricard/nkn-defi/internal/engine/risk"
    "github.com/fabricioricard/nkn-defi/internal/engine/pricing"
    "github.com/fabricioricard/nkn-defi/internal/engine/treasury"
)

type PoolUsecase struct{}
func NewPoolUsecase(poolRepo repo.PoolRepository, cache cache.Cache, bus eventbus.Publisher, amm *amm.ConstantProductAMM) *PoolUsecase {
    return &PoolUsecase{}
}

type SwapUsecase struct{}
func NewSwapUsecase(poolRepo repo.PoolRepository, swapRepo repo.SwapRepository, cache cache.Cache, bus eventbus.Publisher, amm *amm.ConstantProductAMM, risk *risk.Engine, pricing *pricing.Oracle, treasury *treasury.Engine) *SwapUsecase {
    return &SwapUsecase{}
}

type StakingUsecase struct{}
func NewStakeUsecase(stakeRepo repo.StakeRepository, bus eventbus.Publisher) *StakingUsecase {
    return &StakingUsecase{}
}

type NodeUsecase struct{}
func NewNodeUsecase(nodeRepo repo.NodeRepository, bus eventbus.Publisher) *NodeUsecase {
    return &NodeUsecase{}
}

type LendingUsecase struct{}
type RewardUsecase struct{}
type TreasuryUsecase struct{}
