package app

import (
    "context"
    "math/big"
    "time"
    "github.com/google/uuid"
    "github.com/fabricioricard/nkn-defi/internal/domain"
    "github.com/fabricioricard/nkn-defi/internal/ports/cache"
    "github.com/fabricioricard/nkn-defi/internal/ports/eventbus"
    "github.com/fabricioricard/nkn-defi/internal/ports/repo"
)

type PoolUsecase struct {
    repo  repo.PoolRepository
    cache cache.Cache
    bus   eventbus.Publisher
}

func NewPoolUsecase(repo repo.PoolRepository, cache cache.Cache, bus eventbus.Publisher) *PoolUsecase {
    return &PoolUsecase{repo: repo, cache: cache, bus: bus}
}

func (uc *PoolUsecase) CreatePool(ctx context.Context, token0, token1 string, feeBps int) (*domain.Pool, error) {
    pool := &domain.Pool{
        ID:               uuid.New().String(),
        Token0:           token0,
        Token1:           token1,
        Reserve0:         big.NewInt(0),
        Reserve1:         big.NewInt(0),
        TotalLPTokens:    big.NewInt(0),
        FeeBps:           feeBps,
        ProtocolFeeShare: 10,
    }
    if err := uc.repo.Create(ctx, pool); err != nil {
        return nil, err
    }
    return pool, nil
}

func (uc *PoolUsecase) GetPool(ctx context.Context, id string) (*domain.Pool, error) {
    pool := &domain.Pool{}
    if err := uc.cache.Get(ctx, "pool:"+id, pool); err == nil {
        return pool, nil
    }
    pool, err := uc.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    uc.cache.Set(ctx, "pool:"+id, pool, 1*time.Minute)
    return pool, nil
}

func (uc *PoolUsecase) ListPools(ctx context.Context) ([]*domain.Pool, error) {
    return uc.repo.List(ctx)
}

func (uc *PoolUsecase) AddLiquidity(ctx context.Context, poolID, owner string, amount0, amount1 *big.Int) (*domain.Pool, *big.Int, error) {
    pool, err := uc.GetPool(ctx, poolID)
    if err != nil {
        return nil, nil, err
    }
    lp := pool.AddLiquidity(amount0, amount1)
    if err := uc.repo.Update(ctx, pool); err != nil {
        return nil, nil, err
    }
    uc.cache.Set(ctx, "pool:"+poolID, pool, 1*time.Minute)
    uc.bus.Publish("liquidity.added", map[string]interface{}{"poolID": poolID, "owner": owner, "lp": lp.String()})
    return pool, lp, nil
}

func (uc *PoolUsecase) Swap(ctx context.Context, poolID, sender, tokenIn string, amountIn *big.Int) (amountOut *big.Int, fee *big.Int, err error) {
    pool, err := uc.GetPool(ctx, poolID)
    if err != nil {
        return nil, nil, err
    }
    amountOut, fee, err = pool.Swap(tokenIn, amountIn)
    if err != nil {
        return nil, nil, err
    }
    if err := uc.repo.Update(ctx, pool); err != nil {
        return nil, nil, err
    }
    uc.cache.Set(ctx, "pool:"+poolID, pool, 1*time.Minute)
    uc.bus.Publish("swap.executed", map[string]interface{}{"poolID": poolID, "sender": sender, "amountIn": amountIn.String(), "amountOut": amountOut.String()})
    return amountOut, fee, nil
}
