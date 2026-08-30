package sync

import (
	"context"
	"database/sql"
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
	log.Println("BridgeDepositWorker started (balance polling with flexible conversion)")

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

				balance, err := w.nknClient.GetAddressBalance(dep.NknDepositAddress)
				if err != nil {
					log.Printf("error fetching balance for %s: %v", dep.NknDepositAddress, err)
					continue
				}

				if balance.Cmp(big.NewInt(0)) <= 0 {
					log.Printf("no balance yet for deposit %s", dep.ID)
					continue
				}

				// Converte o saldo para wei (18 decimais) de forma flexível.
				// Se o saldo for menor que 10^10, assume que está em NKN inteiro.
				// Caso contrário, assume que está em unidades de 8 decimais.
				threshold := new(big.Int).Exp(big.NewInt(10), big.NewInt(10), nil) // 10^10
				var amountWei *big.Int

				if balance.Cmp(threshold) < 0 {
					// Saldo em NKN (ex: 1 NKN) -> multiplica por 10^18
					multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
					amountWei = new(big.Int).Mul(balance, multiplier)
				} else {
					// Saldo em unidades mínimas (8 decimais) -> multiplica por 10^10
					multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(10), nil)
					amountWei = new(big.Int).Mul(balance, multiplier)
				}

				log.Printf("balance raw: %s, amountWei to mint: %s", balance.String(), amountWei.String())

				var zeroHash [32]byte

				mintTxHash, err := w.ethClient.MintWNKN(
					common.HexToAddress(dep.EthAddress),
					amountWei,
					zeroHash,
				)
				if err != nil {
					log.Printf("failed to mint wNKN for deposit %s: %v", dep.ID, err)
					continue
				}

				err = w.bridgeRepo.UpdateDepositAfterMint(ctx, dep.ID, "auto-balance", mintTxHash, "completed")
				if err != nil {
					log.Printf("failed to update deposit %s: %v", dep.ID, err)
				} else {
					log.Printf("Minted %s wNKN to %s (deposit %s, mint tx %s)", amountWei.String(), dep.EthAddress, dep.ID, mintTxHash)
				}
			}
		}
	}
}