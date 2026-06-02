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

	// Pega o bloco atual para começar a monitorar daqui em diante
	currentBlock, err := w.ethClient.BlockNumber(ctx)
	if err != nil {
		log.Printf("failed to get initial block: %v", err)
		currentBlock = 0
	}
	lastBlock := int64(currentBlock)
	log.Printf("BridgeWithdrawWorker started, monitoring from block %d", lastBlock)

	for {
		select {
		case <-ctx.Done():
			log.Println("BridgeWithdrawWorker stopped")
			return
		case <-ticker.C:
			newBlock, err := w.ethClient.BlockNumber(ctx)
			if err != nil {
				log.Printf("failed to get block number: %v", err)
				continue
			}
			if int64(newBlock) > lastBlock {
				// Filtra eventos apenas no intervalo (lastBlock, newBlock]
				events, err := w.ethClient.FilterBurnEvents(ctx, lastBlock)
				if err != nil {
					log.Printf("error filtering burn events: %v", err)
					continue
				}
				for _, event := range events {
					w.processBurnEvent(ctx, event)
				}
				lastBlock = int64(newBlock) + 1
			}
		}
	}
}

func (w *BridgeWithdrawWorker) processBurnEvent(ctx context.Context, event *ethereum.WrappedNKNBurnRequested) {
	burnTxHash := event.Raw.TxHash.Hex()
	log.Printf("BurnRequested: from=%s, amount=%s, nknAddress=%s, tx=%s",
		event.From.Hex(), event.Amount.String(), event.NknMainnetAddress, burnTxHash)

	existing, _ := w.bridgeRepo.GetWithdrawalByBurnTxHash(ctx, burnTxHash)
	if existing != nil {
		log.Printf("burn tx %s already processed, skipping", burnTxHash)
		return
	}

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

	amount := event.Amount
	fee := new(big.Int).Div(amount, big.NewInt(1000))
	sendAmount := new(big.Int).Sub(amount, fee)

	if w.nknSender == nil {
		log.Printf("NKN wallet not available, skipping send for tx %s", burnTxHash)
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
		log.Printf("Withdrawal completed: sent %s NKN to %s (tx %s)", sendAmount.String(), event.NknMainnetAddress, mainnetTxHash)
	}
}