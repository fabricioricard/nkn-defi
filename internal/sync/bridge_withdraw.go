package sync

import (
	"context"
	"database/sql"
	"log"
	"math/big"
	"os"
	"time"

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
	nknSender, err := nknclient.NewWallet(os.Getenv("BRIDGE_NKN_PRIVATE_KEY"))
	if err != nil {
		log.Printf("WARNING: could not create NKN wallet: %v (withdrawals will be skipped)", err)
		nknSender = nil
	}
	return &BridgeWithdrawWorker{
		db:         db,
		nknClient:  nkn,
		ethClient:  eth,
		nknSender:  nknSender,
		bridgeRepo: bridgeRepo,
	}
}

func (w *BridgeWithdrawWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// Bloco inicial: busca o último bloco processado salvo, ou 0
	lastBlock := int64(0) // idealmente você salvaria no banco para não reprocessar tudo
	log.Println("BridgeWithdrawWorker started, polling for BurnRequested events")

	for {
		select {
		case <-ctx.Done():
			log.Println("BridgeWithdrawWorker stopped")
			return
		case <-ticker.C:
			// Obtém o bloco atual
			currentBlock, err := w.ethClient.BlockNumber(ctx)
			if err != nil {
				log.Printf("failed to get block number: %v", err)
				continue
			}

			if int64(currentBlock) > lastBlock {
				events, err := w.ethClient.FilterBurnEvents(ctx, lastBlock)
				if err != nil {
					log.Printf("error filtering burn events: %v", err)
					continue
				}
				for _, event := range events {
					w.processBurnEvent(ctx, event)
				}
				// Atualiza o bloco processado para o próximo (evita reprocessar os mesmos eventos)
				lastBlock = int64(currentBlock) + 1
			}
		}
	}
}

func (w *BridgeWithdrawWorker) processBurnEvent(ctx context.Context, event *ethereum.WrappedNKNBurnRequested) {
	burnTxHash := event.Raw.TxHash.Hex()
	log.Printf("BurnRequested received: from=%s, amount=%s, nknAddress=%s, tx=%s",
		event.From.Hex(), event.Amount.String(), event.NknMainnetAddress, burnTxHash)

	// Verifica se já processamos
	existing, _ := w.bridgeRepo.GetWithdrawalByBurnTxHash(ctx, burnTxHash)
	if existing != nil {
		log.Printf("burn tx %s already processed, skipping", burnTxHash)
		return
	}

	// Insere retirada como pending
	err := w.bridgeRepo.InsertWithdrawal(ctx, &postgres.BridgeWithdrawal{
		ID:          "",
		BurnTxHash:  burnTxHash,
		FromAddress: event.From.Hex(),
		Amount:      event.Amount.String(),
		NknAddress:  event.NknMainnetAddress,
		Status:      "pending",
	})
	if err != nil {
		log.Printf("failed to insert withdrawal: %v", err)
		return
	}

	// Calcula taxa (0.1%)
	amount := event.Amount
	fee := new(big.Int).Div(amount, big.NewInt(1000))
	sendAmount := new(big.Int).Sub(amount, fee)

	if w.nknSender == nil {
		log.Printf("WARNING: NKN wallet not available, skipping send for burn tx %s", burnTxHash)
		_ = w.bridgeRepo.UpdateWithdrawalAfterSend(ctx, burnTxHash, "", "pending_manual")
		return
	}

	mainnetTxHash, err := w.nknSender.SendNKN(event.NknMainnetAddress, sendAmount)
	if err != nil {
		log.Printf("failed to send NKN withdrawal: %v", err)
		_ = w.bridgeRepo.UpdateWithdrawalAfterSend(ctx, burnTxHash, "", "failed")
		return
	}

	err = w.bridgeRepo.UpdateWithdrawalAfterSend(ctx, burnTxHash, mainnetTxHash, "completed")
	if err != nil {
		log.Printf("failed to update withdrawal: %v", err)
	} else {
		log.Printf("Withdrawal completed: sent %s NKN to %s (mainnet tx %s)", sendAmount.String(), event.NknMainnetAddress, mainnetTxHash)
	}
}