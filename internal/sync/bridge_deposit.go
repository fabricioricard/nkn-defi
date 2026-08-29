package sync

import (
	"context"
	"database/sql"
	"encoding/hex"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fabricioricard/nkn-defi/internal/adapters/ethereum"
	"github.com/fabricioricard/nkn-defi/internal/adapters/nknclient"
	"github.com/fabricioricard/nkn-defi/internal/adapters/postgres"
)

type BridgeDepositWorker struct {
	db         *sql.DB
	nknClient  *nknclient.Client
	ethClient  *ethereum.Client
	bridgeRepo *postgres.BridgeRepository
}

func NewBridgeDepositWorker(
	db *sql.DB,
	nkn *nknclient.Client,
	eth *ethereum.Client,
	bridgeRepo *postgres.BridgeRepository,
) *BridgeDepositWorker {
	return &BridgeDepositWorker{
		db:         db,
		nknClient:  nkn,
		ethClient:  eth,
		bridgeRepo: bridgeRepo,
	}
}

func (w *BridgeDepositWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	log.Println("BridgeDepositWorker started (monitoring pending deposits via transactions)")

	for {
		select {
		case <-ctx.Done():
			log.Println("BridgeDepositWorker stopped")
			return
		case <-ticker.C:
			deposits, err := w.bridgeRepo.ListPendingDeposits(ctx)
			if err != nil {
				log.Printf("error listing pending deposits: %v", err)
				continue
			}
			for _, dep := range deposits {
				log.Printf("checking deposit %s at address %s", dep.ID, dep.NknDepositAddress)

				txs, err := w.nknClient.GetRecentTransactions(dep.NknDepositAddress)
				if err != nil {
					log.Printf("error fetching txs for %s: %v", dep.NknDepositAddress, err)
					continue
				}

				for _, tx := range txs {
					if tx.Amount == nil || tx.Amount.Cmp(big.NewInt(0)) <= 0 {
						continue
					}

					// Processa primeira transação com valor positivo
					txHashBytes, err := hexStringToBytes32(tx.Hash)
					if err != nil {
						log.Printf("invalid tx hash %s: %v", tx.Hash, err)
						continue
					}

					mintTxHash, err := w.ethClient.MintWNKN(
						common.HexToAddress(dep.EthAddress),
						tx.Amount,
						txHashBytes,
					)
					if err != nil {
						log.Printf("failed to mint wNKN for deposit %s: %v", dep.ID, err)
						continue
					}

					err = w.bridgeRepo.UpdateDepositAfterMint(ctx, dep.ID, tx.Hash, mintTxHash, "completed")
					if err != nil {
						log.Printf("failed to update deposit %s: %v", dep.ID, err)
					} else {
						log.Printf("Minted %s wNKN to %s (deposit %s, mint tx %s)", tx.Amount.String(), dep.EthAddress, dep.ID, mintTxHash)
					}
					break // sai após processar a primeira transação válida
				}
			}
		}
	}
}

// hexStringToBytes32 converte string hex para [32]byte.
func hexStringToBytes32(hexStr string) ([32]byte, error) {
	var b [32]byte
	clean := hexStr
	if len(clean) >= 2 && clean[:2] == "0x" {
		clean = clean[2:]
	}
	bytes, err := hex.DecodeString(clean)
	if err != nil {
		return b, err
	}
	copy(b[:], bytes)
	return b, nil
}