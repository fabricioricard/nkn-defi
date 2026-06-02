package sync

import (
	"context"
	"database/sql"
	"log"
	"math/big"
	"os"

	"github.com/fabricioricard/nkn-defi/internal/adapters/ethereum"
	"github.com/fabricioricard/nkn-defi/internal/adapters/nknclient"
	"github.com/fabricioricard/nkn-defi/internal/adapters/postgres"
)

type BridgeWithdrawWorker struct {
	db         *sql.DB
	nknClient  *nknclient.Client
	ethClient  *ethereum.Client
	nknSender  *nknclient.Wallet
	bridgeRepo *postgres.BridgeRepository
}

func NewBridgeWithdrawWorker(
	db *sql.DB,
	nkn *nknclient.Client,
	eth *ethereum.Client,
	bridgeRepo *postgres.BridgeRepository,
) *BridgeWithdrawWorker {
	nknSender := nknclient.NewWallet(os.Getenv("BRIDGE_NKN_PRIVATE_KEY"))
	return &BridgeWithdrawWorker{
		db:         db,
		nknClient:  nkn,
		ethClient:  eth,
		nknSender:  nknSender,
		bridgeRepo: bridgeRepo,
	}
}

func (w *BridgeWithdrawWorker) Start(ctx context.Context) {
	ch, err := w.ethClient.ListenBurnEvents(ctx)
	if err != nil {
		log.Fatalf("failed to listen burn events: %v", err)
	}
	log.Println("BridgeWithdrawWorker started, listening for BurnRequested events")

	for {
		select {
		case <-ctx.Done():
			log.Println("BridgeWithdrawWorker stopped")
			return
		case event := <-ch:
			burnTxHash := event.Raw.TxHash.Hex()
			log.Printf("BurnRequested received: from=%s, amount=%s, nknAddress=%s, tx=%s",
				event.From.Hex(), event.Amount.String(), event.NknMainnetAddress, burnTxHash)

			// Verifica se já processamos essa transação de burn
			existing, _ := w.bridgeRepo.GetWithdrawalByBurnTxHash(ctx, burnTxHash)
			if existing != nil {
				log.Printf("burn tx %s already processed, skipping", burnTxHash)
				continue
			}

			// Insere retirada no banco como pending
			err := w.bridgeRepo.InsertWithdrawal(ctx, &postgres.BridgeWithdrawal{
				ID:          "", // auto-generate UUID
				BurnTxHash:  burnTxHash,
				FromAddress: event.From.Hex(),
				Amount:      event.Amount.String(),
				NknAddress:  event.NknMainnetAddress,
				Status:      "pending",
			})
			if err != nil {
				log.Printf("failed to insert withdrawal: %v", err)
				continue
			}

			// Calcula taxa (0.1%)
			amount := event.Amount
			fee := new(big.Int).Div(amount, big.NewInt(1000)) // 0.1%
			sendAmount := new(big.Int).Sub(amount, fee)

			// Envia NKN na mainnet
			mainnetTxHash, err := w.nknSender.SendNKN(event.NknMainnetAddress, sendAmount)
			if err != nil {
				log.Printf("failed to send NKN withdrawal: %v", err)
				// Marca como failed
				_ = w.bridgeRepo.UpdateWithdrawalAfterSend(ctx, burnTxHash, "", "failed")
				continue
			}

			// Atualiza retirada como completed
			err = w.bridgeRepo.UpdateWithdrawalAfterSend(ctx, burnTxHash, mainnetTxHash, "completed")
			if err != nil {
				log.Printf("failed to update withdrawal: %v", err)
			} else {
				log.Printf("Withdrawal completed: sent %s NKN to %s (mainnet tx %s)", sendAmount.String(), event.NknMainnetAddress, mainnetTxHash)
			}
		}
	}
}