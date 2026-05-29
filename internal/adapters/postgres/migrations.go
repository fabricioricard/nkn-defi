package postgres

import "database/sql"

func RunMigrations(db *sql.DB) error {
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS pools (
            id UUID PRIMARY KEY,
            token0 TEXT NOT NULL,
            token1 TEXT NOT NULL,
            reserve0 TEXT NOT NULL,
            reserve1 TEXT NOT NULL,
            total_lp_tokens TEXT NOT NULL,
            fee_bps INT NOT NULL,
            protocol_fee_share INT NOT NULL DEFAULT 10,
            created_at TIMESTAMPTZ DEFAULT NOW(),
            updated_at TIMESTAMPTZ DEFAULT NOW()
        );
    `)
    return err
}