package domain

import (
    "math/big"
    "errors"
)

type Pool struct {
    ID               string   `json:"id"`
    Token0           string   `json:"token0"`
    Token1           string   `json:"token1"`
    Reserve0         *big.Int `json:"reserve0"`
    Reserve1         *big.Int `json:"reserve1"`
    TotalLPTokens    *big.Int `json:"total_lp_tokens"`
    FeeBps           int      `json:"fee_bps"`
    ProtocolFeeShare int      `json:"protocol_fee_share"`
}

// AddLiquidity returns the amount of LP tokens minted.
func (p *Pool) AddLiquidity(amount0, amount1 *big.Int) *big.Int {
    if p.TotalLPTokens.Cmp(big.NewInt(0)) == 0 {
        // Initial mint: geometric mean
        lp := new(big.Int).Sqrt(new(big.Int).Mul(amount0, amount1))
        p.Reserve0 = amount0
        p.Reserve1 = amount1
        p.TotalLPTokens = lp
        return lp
    }
    share0 := new(big.Int).Div(new(big.Int).Mul(amount0, p.TotalLPTokens), p.Reserve0)
    share1 := new(big.Int).Div(new(big.Int).Mul(amount1, p.TotalLPTokens), p.Reserve1)
    mint := share0
    if share1.Cmp(share0) < 0 {
        mint = share1
    }
    p.Reserve0.Add(p.Reserve0, amount0)
    p.Reserve1.Add(p.Reserve1, amount1)
    p.TotalLPTokens.Add(p.TotalLPTokens, mint)
    return mint
}

// RemoveLiquidity burns LP tokens and returns amounts.
func (p *Pool) RemoveLiquidity(lpTokens *big.Int) (amount0, amount1 *big.Int) {
    share0 := new(big.Int).Div(new(big.Int).Mul(lpTokens, p.Reserve0), p.TotalLPTokens)
    share1 := new(big.Int).Div(new(big.Int).Mul(lpTokens, p.Reserve1), p.TotalLPTokens)
    p.Reserve0.Sub(p.Reserve0, share0)
    p.Reserve1.Sub(p.Reserve1, share1)
    p.TotalLPTokens.Sub(p.TotalLPTokens, lpTokens)
    return share0, share1
}

// Swap executes a swap, returning output amount and fee.
func (p *Pool) Swap(tokenIn string, amountIn *big.Int) (amountOut, fee *big.Int, err error) {
    if amountIn.Cmp(big.NewInt(0)) <= 0 {
        return nil, nil, errors.New("amount must be positive")
    }
    var reserveIn, reserveOut *big.Int
    if tokenIn == p.Token0 {
        reserveIn = p.Reserve0
        reserveOut = p.Reserve1
    } else if tokenIn == p.Token1 {
        reserveIn = p.Reserve1
        reserveOut = p.Reserve0
    } else {
        return nil, nil, errors.New("invalid token")
    }

    feeBps := big.NewInt(int64(p.FeeBps))
    fee = new(big.Int).Mul(amountIn, feeBps)
    fee.Div(fee, big.NewInt(10000))
    amountInAfterFee := new(big.Int).Sub(amountIn, fee)

    // k = reserveIn * reserveOut
    numerator := new(big.Int).Mul(amountInAfterFee, reserveOut)
    denominator := new(big.Int).Add(reserveIn, amountInAfterFee)
    amountOut = new(big.Int).Div(numerator, denominator)

    // Update reserves
    reserveIn.Add(reserveIn, amountInAfterFee)
    reserveOut.Sub(reserveOut, amountOut)
    return amountOut, fee, nil
}

// Price returns exchange rate as float64 (approximate).
func (p *Pool) Price() float64 {
    if p.Reserve0.Cmp(big.NewInt(0)) == 0 || p.Reserve1.Cmp(big.NewInt(0)) == 0 {
        return 0
    }
    r0 := new(big.Float).SetInt(p.Reserve0)
    r1 := new(big.Float).SetInt(p.Reserve1)
    price, _ := new(big.Float).Quo(r1, r0).Float64() // token1 per token0
    return price
}