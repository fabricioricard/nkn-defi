package postgres

import "database/sql"

func RunMigrations(db *sql.DB) error {
	// Garantir que a extensão uuid-ossp esteja disponível
	if _, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`); err != nil {
		return err
	}

	// Tabela de pools
	if _, err := db.Exec(`
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
	`); err != nil {
		return err
	}

	// Tabela de depósitos da Bridge
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS bridge_deposits (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			eth_address TEXT NOT NULL,
			amount NUMERIC NOT NULL,
			nkn_deposit_address TEXT NOT NULL,
			status TEXT DEFAULT 'pending',
			mainnet_tx_hash TEXT,
			mint_tx_hash TEXT,
			memo TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);
	`); err != nil {
		return err
	}

	// Garantir que a coluna memo exista (para tabelas criadas antes desta migração)
	if _, err := db.Exec(`ALTER TABLE bridge_deposits ADD COLUMN IF NOT EXISTS memo TEXT NOT NULL DEFAULT '';`); err != nil {
		return err
	}

	// Tabela de retiradas da Bridge
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS bridge_withdrawals (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			burn_tx_hash TEXT NOT NULL UNIQUE,
			from_address TEXT NOT NULL,
			amount NUMERIC NOT NULL,
			nkn_address TEXT NOT NULL,
			status TEXT DEFAULT 'pending',
			mainnet_tx_hash TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);
	`); err != nil {
		return err
	}

	return nil
}