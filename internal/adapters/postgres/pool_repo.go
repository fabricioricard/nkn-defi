package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	return db, nil
}

// Stubs de repositórios (retornam nil – depois implemente os verdadeiros)
func NewPoolRepository(db *sql.DB) *PoolRepository       { return &PoolRepository{} }
func NewSwapRepository(db *sql.DB) *SwapRepository       { return &SwapRepository{} }
func NewStakeRepository(db *sql.DB) *StakeRepository     { return &StakeRepository{} }
func NewTreasuryRepository(db *sql.DB) *TreasuryRepository { return &TreasuryRepository{} }
func NewNodeRepository(db *sql.DB) *NodeRepository       { return &NodeRepository{} }

type PoolRepository struct{}
type SwapRepository struct{}
type StakeRepository struct{}
type TreasuryRepository struct{}
type NodeRepository struct{}