package repo

import "context"

type PoolRepository interface {
}

type SwapRepository interface {
}

type StakeRepository interface {
}

type TreasuryRepository interface {
}

type NodeRepository interface {
}

// Apenas para evitar erro de "imported and not used" se nenhum método usar context
var _ = context.Background
