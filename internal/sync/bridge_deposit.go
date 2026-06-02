package sync

import (
	"context"
	"database/sql"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fabricioricard/nkn-defi/internal/adapters/ethereum"
	"github.com/fabricioricard/nkn-defi/internal/adapters/nknclient"
	"github.com/fabricioricard/nkn-defi/internal/adapters/postgres"
)

type BridgeDepositWorker struct {
	db          *sql.DB
	nknClient   *nknclient.Client
	ethClient   *ethereum.Client
	depositAddr string
	bridgeRepo  *postgres.BridgeRepository
}

func NewBridgeDepositWorker(
	db *sql.DB,
	nkn *nknclient.Client,
	eth *ethereum.Client,
	bridgeRepo *postgres.BridgeRepository,
) *BridgeDepositWorker {
	return &BridgeDepositWorker{
		db:          db,
		nknClient:   nkn,
		ethClient:   eth,
		depositAddr: os.Getenv("BRIDGE_NKN_ADDRESS"),
		bridgeRepo:  bridgeRepo,
	}
}

func (w *BridgeDepositWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	log.Printf("BridgeDepositWorker started, watching address %s", w.depositAddr)

	for {
		select {
		case <-ctx.Done():
			log.Println("BridgeDepositWorker stopped")
			return
		case <-ticker.C:
			txs, err := w.nknClient.GetRecentTransactions(w.depositAddr)
			if err != nil {
				log.Printf("error fetching NKN txs: %v", err)
				continue
			}
			for _, tx := range txs {
				if tx.Amount == nil || tx.Amount.Cmp(big.NewInt(0)) <= 0 {
					continue
				}
				// Extrai o memo (campo de dados da transação NKN)
				memo := tx.Memo // ajuste conforme a struct real de transação da API NKN

				// Procura depósito pendente com esse memo
				dep, err := w.bridgeRepo.FindPendingDepositByMemo(ctx, memo)
				if err != nil || dep == nil {
					log.Printf("no pending deposit for memo %s: %v", memo, err)
					continue
				}

				// Converte o hash da transação mainnet para [32]byte
				txHashBytes, err := hexToBytes32(tx.Hash)
				if err != nil {
					log.Printf("invalid tx hash %s: %v", tx.Hash, err)
					continue
				}

				// Mint wNKN para o endereço Ethereum do depósito
				mintTxHash, err := w.ethClient.MintWNKN(
					common.HexToAddress(dep.EthAddress),
					tx.Amount,
					txHashBytes,
				)
				if err != nil {
					log.Printf("failed to mint wNKN for deposit %s: %v", dep.ID, err)
					continue
				}

				// Atualiza o depósito no banco
				err = w.bridgeRepo.UpdateDepositAfterMint(ctx, dep.ID, tx.Hash, mintTxHash, "completed")
				if err != nil {
					log.Printf("failed to update deposit %s: %v", dep.ID, err)
				} else {
					log.Printf("Minted %s wNKN to %s (deposit %s, mint tx %s)", tx.Amount.String(), dep.EthAddress, dep.ID, mintTxHash)
				}
			}
		}
	}
}

// hexToBytes32 converte uma string hexadecimal (com ou sem 0x) para um array de 32 bytes.
func hexToBytes32(hex string) ([32]byte, error) {
	var b [32]byte
	// remove prefixo "0x" se existir
	if len(hex) >= 2 && hex[:2] == "0x" {
		hex = hex[2:]
	}
	// decodifica
	bytes, err := hex.DecodeString(hex) // precisa import "encoding/hex"
	if err != nil {
		return b, err
	}
	copy(b[:], bytes)
	return b, nil
}