package postgres

import (
    "context"
    "database/sql"
    "fmt"
    "math/big"

    _ "github.com/lib/pq"

    "github.com/fabricioricard/nkn-defi/internal/domain"
)

// Connect abre a conexão com o banco
func Connect(databaseURL string) (*sql.DB, error) {
    db, err := sql.Open("postgres", databaseURL)
    if err != nil {
        return nil, fmt.Errorf("failed to open db: %w", err)
    }
    return db, nil
}

type poolRepo struct {
    db *sql.DB
}

func NewPoolRepository(db *sql.DB) *poolRepo {
    return &poolRepo{db: db}
}

func (r *poolRepo) Create(ctx context.Context, p *domain.Pool) error {
    _, err := r.db.ExecContext(ctx, `
        INSERT INTO pools (id, token0, token1, reserve0, reserve1, total_lp_tokens, fee_bps, protocol_fee_share)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
        p.ID, p.Token0, p.Token1, p.Reserve0.String(), p.Reserve1.String(), p.TotalLPTokens.String(), p.FeeBps, p.ProtocolFeeShare)
    return err
}

func (r *poolRepo) GetByID(ctx context.Context, id string) (*domain.Pool, error) {
    p := &domain.Pool{}
    var r0, r1, total string
    err := r.db.QueryRowContext(ctx, `SELECT id, token0, token1, reserve0, reserve1, total_lp_tokens, fee_bps, protocol_fee_share FROM pools WHERE id=$1`, id).
        Scan(&p.ID, &p.Token0, &p.Token1, &r0, &r1, &total, &p.FeeBps, &p.ProtocolFeeShare)
    if err != nil {
        return nil, err
    }
    p.Reserve0, _ = new(big.Int).SetString(r0, 10)
    p.Reserve1, _ = new(big.Int).SetString(r1, 10)
    p.TotalLPTokens, _ = new(big.Int).SetString(total, 10)
    return p, nil
}

func (r *poolRepo) List(ctx context.Context) ([]*domain.Pool, error) {
    rows, err := r.db.QueryContext(ctx, `SELECT id, token0, token1, reserve0, reserve1, total_lp_tokens, fee_bps, protocol_fee_share FROM pools`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var pools []*domain.Pool
    for rows.Next() {
        p := &domain.Pool{}
        var r0, r1, total string
        if err := rows.Scan(&p.ID, &p.Token0, &p.Token1, &r0, &r1, &total, &p.FeeBps, &p.ProtocolFeeShare); err != nil {
            return nil, err
        }
        p.Reserve0, _ = new(big.Int).SetString(r0, 10)
        p.Reserve1, _ = new(big.Int).SetString(r1, 10)
        p.TotalLPTokens, _ = new(big.Int).SetString(total, 10)
        pools = append(pools, p)
    }
    return pools, nil
}

func (r *poolRepo) Update(ctx context.Context, p *domain.Pool) error {
    _, err := r.db.ExecContext(ctx, `UPDATE pools SET reserve0=$1, reserve1=$2, total_lp_tokens=$3, fee_bps=$4 WHERE id=$5`,
        p.Reserve0.String(), p.Reserve1.String(), p.TotalLPTokens.String(), p.FeeBps, p.ID)
    return err
}