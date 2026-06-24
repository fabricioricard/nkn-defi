package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// BridgeDeposit representa um depósito pendente ou concluído.
type BridgeDeposit struct {
	ID                 string    `json:"id"`
	EthAddress         string    `json:"eth_address"`
	Amount             string    `json:"amount"`
	NknDepositAddress  string    `json:"nkn_deposit_address"`
	Status             string    `json:"status"`
	MainnetTxHash      string    `json:"mainnet_tx_hash,omitempty"`
	MintTxHash         string    `json:"mint_tx_hash,omitempty"`
	Memo               string    `json:"memo"`
	NknDerivedIndex    int64     `json:"nkn_derived_index"` // índice usado para derivar o endereço
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// BridgeWithdrawal representa uma retirada em andamento.
type BridgeWithdrawal struct {
	ID            string    `json:"id"`
	BurnTxHash    string    `json:"burn_tx_hash"`
	FromAddress   string    `json:"from_address"`
	Amount        string    `json:"amount"`
	NknAddress    string    `json:"nkn_address"`
	Status        string    `json:"status"`
	MainnetTxHash string    `json:"mainnet_tx_hash,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// BridgeRepository gerencia as operações de banco da Bridge.
type BridgeRepository struct {
	db *sql.DB
}

func NewBridgeRepository(db *sql.DB) *BridgeRepository {
	return &BridgeRepository{db: db}
}

// InsertDeposit insere um novo depósito no banco.
func (r *BridgeRepository) InsertDeposit(ctx context.Context, deposit *BridgeDeposit) error {
	if deposit.ID == "" {
		deposit.ID = uuid.New().String()
	}
	deposit.CreatedAt = time.Now()
	deposit.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO bridge_deposits (id, eth_address, amount, nkn_deposit_address, status, memo, nkn_derived_index, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		deposit.ID, deposit.EthAddress, deposit.Amount, deposit.NknDepositAddress,
		deposit.Status, deposit.Memo, deposit.NknDerivedIndex, deposit.CreatedAt, deposit.UpdatedAt,
	)
	return err
}

// GetDepositByID busca um depósito pelo ID.
func (r *BridgeRepository) GetDepositByID(ctx context.Context, id string) (*BridgeDeposit, error) {
	dep := &BridgeDeposit{}
	var mainnetTxHash, mintTxHash sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, eth_address, amount, nkn_deposit_address, status, mainnet_tx_hash, mint_tx_hash, memo, nkn_derived_index, created_at, updated_at
		FROM bridge_deposits WHERE id = $1`, id).
		Scan(&dep.ID, &dep.EthAddress, &dep.Amount, &dep.NknDepositAddress,
			&dep.Status, &mainnetTxHash, &mintTxHash, &dep.Memo, &dep.NknDerivedIndex, &dep.CreatedAt, &dep.UpdatedAt)
	if err != nil {
		return nil, err
	}
	dep.MainnetTxHash = mainnetTxHash.String
	dep.MintTxHash = mintTxHash.String
	return dep, nil
}

// FindPendingDepositByMemo busca um depósito pendente com base no memo.
func (r *BridgeRepository) FindPendingDepositByMemo(ctx context.Context, memo string) (*BridgeDeposit, error) {
	dep := &BridgeDeposit{}
	var mainnetTxHash, mintTxHash sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, eth_address, amount, nkn_deposit_address, status, mainnet_tx_hash, mint_tx_hash, memo, nkn_derived_index, created_at, updated_at
		FROM bridge_deposits WHERE memo = $1 AND status = 'pending'`, memo).
		Scan(&dep.ID, &dep.EthAddress, &dep.Amount, &dep.NknDepositAddress,
			&dep.Status, &mainnetTxHash, &mintTxHash, &dep.Memo, &dep.NknDerivedIndex, &dep.CreatedAt, &dep.UpdatedAt)
	if err != nil {
		return nil, err
	}
	dep.MainnetTxHash = mainnetTxHash.String
	dep.MintTxHash = mintTxHash.String
	return dep, nil
}

// UpdateDepositAfterMint atualiza o depósito após o mint.
func (r *BridgeRepository) UpdateDepositAfterMint(ctx context.Context, id, mainnetTxHash, mintTxHash, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE bridge_deposits SET mainnet_tx_hash = $1, mint_tx_hash = $2, status = $3, updated_at = NOW()
		WHERE id = $4`,
		mainnetTxHash, mintTxHash, status, id,
	)
	return err
}

// InsertWithdrawal insere uma nova retirada.
func (r *BridgeRepository) InsertWithdrawal(ctx context.Context, w *BridgeWithdrawal) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	w.CreatedAt = time.Now()
	w.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO bridge_withdrawals (id, burn_tx_hash, from_address, amount, nkn_address, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		w.ID, w.BurnTxHash, w.FromAddress, w.Amount, w.NknAddress,
		w.Status, w.CreatedAt, w.UpdatedAt,
	)
	return err
}

// GetWithdrawalByBurnTxHash busca uma retirada pelo hash da transação de burn.
func (r *BridgeRepository) GetWithdrawalByBurnTxHash(ctx context.Context, burnTxHash string) (*BridgeWithdrawal, error) {
	w := &BridgeWithdrawal{}
	var mainnetTxHash sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, burn_tx_hash, from_address, amount, nkn_address, status, mainnet_tx_hash, created_at, updated_at
		FROM bridge_withdrawals WHERE burn_tx_hash = $1`, burnTxHash).
		Scan(&w.ID, &w.BurnTxHash, &w.FromAddress, &w.Amount, &w.NknAddress,
			&w.Status, &mainnetTxHash, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	w.MainnetTxHash = mainnetTxHash.String
	return w, nil
}

// UpdateWithdrawalAfterSend atualiza a retirada com o hash da transação mainnet e status.
func (r *BridgeRepository) UpdateWithdrawalAfterSend(ctx context.Context, burnTxHash, mainnetTxHash, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE bridge_withdrawals SET mainnet_tx_hash = $1, status = $2, updated_at = NOW()
		WHERE burn_tx_hash = $3`,
		mainnetTxHash, status, burnTxHash,
	)
	return err
}

// ListDepositsByEthAddress lista depósitos filtrados por endereço Ethereum.
func (r *BridgeRepository) ListDepositsByEthAddress(ctx context.Context, ethAddress string) ([]*BridgeDeposit, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, eth_address, amount, nkn_deposit_address, status, mainnet_tx_hash, mint_tx_hash, memo, nkn_derived_index, created_at, updated_at
		FROM bridge_deposits WHERE eth_address = $1 ORDER BY created_at DESC LIMIT 50`, ethAddress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var deposits []*BridgeDeposit
	for rows.Next() {
		dep := &BridgeDeposit{}
		var mainnetTxHash, mintTxHash sql.NullString
		if err := rows.Scan(&dep.ID, &dep.EthAddress, &dep.Amount, &dep.NknDepositAddress,
			&dep.Status, &mainnetTxHash, &mintTxHash, &dep.Memo, &dep.NknDerivedIndex, &dep.CreatedAt, &dep.UpdatedAt); err != nil {
			return nil, err
		}
		dep.MainnetTxHash = mainnetTxHash.String
		dep.MintTxHash = mintTxHash.String
		deposits = append(deposits, dep)
	}
	return deposits, nil
}

// ListPendingDeposits retorna todos os depósitos com status 'pending'.
func (r *BridgeRepository) ListPendingDeposits(ctx context.Context) ([]*BridgeDeposit, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, eth_address, amount, nkn_deposit_address, status, mainnet_tx_hash, mint_tx_hash, memo, nkn_derived_index, created_at, updated_at
		FROM bridge_deposits WHERE status = 'pending'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var deposits []*BridgeDeposit
	for rows.Next() {
		dep := &BridgeDeposit{}
		var mainnetTxHash, mintTxHash sql.NullString
		if err := rows.Scan(&dep.ID, &dep.EthAddress, &dep.Amount, &dep.NknDepositAddress,
			&dep.Status, &mainnetTxHash, &mintTxHash, &dep.Memo, &dep.NknDerivedIndex, &dep.CreatedAt, &dep.UpdatedAt); err != nil {
			return nil, err
		}
		dep.MainnetTxHash = mainnetTxHash.String
		dep.MintTxHash = mintTxHash.String
		deposits = append(deposits, dep)
	}
	return deposits, nil
}

// ListWithdrawalsByFromAddress lista retiradas filtradas por endereço Ethereum de origem.
func (r *BridgeRepository) ListWithdrawalsByFromAddress(ctx context.Context, fromAddress string) ([]*BridgeWithdrawal, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, burn_tx_hash, from_address, amount, nkn_address, status, mainnet_tx_hash, created_at, updated_at
		FROM bridge_withdrawals WHERE from_address = $1 ORDER BY created_at DESC LIMIT 50`, fromAddress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var withdrawals []*BridgeWithdrawal
	for rows.Next() {
		w := &BridgeWithdrawal{}
		var mainnetTxHash sql.NullString
		if err := rows.Scan(&w.ID, &w.BurnTxHash, &w.FromAddress, &w.Amount, &w.NknAddress,
			&w.Status, &mainnetTxHash, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		w.MainnetTxHash = mainnetTxHash.String
		withdrawals = append(withdrawals, w)
	}
	return withdrawals, nil
}