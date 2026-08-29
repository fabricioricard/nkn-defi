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
	log.Println("BridgeDepositWorker started (monitoring pending deposits by balance)")

	for {
		select {
		case <-ctx.Done():
			log.Println("BridgeDepositWorker stopped")
			return
		case <-ticker.C:
			// Lista todos os depósitos pendentes
			deposits, err := w.bridgeRepo.ListPendingDeposits(ctx)
			if err != nil {
				log.Printf("error listing pending deposits: %v", err)
				continue
			}
			for _, dep := range deposits {
				log.Printf("checking deposit %s at address %s", dep.ID, dep.NknDepositAddress)

				// Consulta o saldo do endereço derivado
				balance, err := w.nknClient.GetAddressBalance(dep.NknDepositAddress)
				if err != nil {
					log.Printf("error fetching balance for %s: %v", dep.NknDepositAddress, err)
					continue
				}
				if balance.Cmp(big.NewInt(0)) <= 0 {
					log.Printf("no balance yet for deposit %s", dep.ID)
					continue
				}

				// Converte o valor esperado para *big.Int (opcional, para manter consistência)
				amountWei, ok := new(big.Int).SetString(dep.Amount, 10)
				if !ok {
					// Se falhar, usa o saldo real como valor a cunhar
					amountWei = balance
				}

				// Usa hash zero (bytes32 vazio) porque não temos o hash da transação
				var zeroHash [32]byte

				// Realiza o mint
				mintTxHash, err := w.ethClient.MintWNKN(
					common.HexToAddress(dep.EthAddress),
					amountWei,
					zeroHash,
				)
				if err != nil {
					log.Printf("failed to mint wNKN for deposit %s: %v", dep.ID, err)
					continue
				}

				// Atualiza o depósito como concluído (usa hash dummy para mainnetTxHash)
				err = w.bridgeRepo.UpdateDepositAfterMint(ctx, dep.ID, "auto-balance", mintTxHash, "completed")
				if err != nil {
					log.Printf("failed to update deposit %s: %v", dep.ID, err)
				} else {
					log.Printf("Minted %s wNKN to %s (deposit %s, mint tx %s)", amountWei.String(), dep.EthAddress, dep.ID, mintTxHash)
				}
				// Sai do loop do depósito atual para evitar mint duplicado
				break
			}
		}
	}
}