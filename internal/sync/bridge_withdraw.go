package sync

import (
	"context"
	"log"
	"math/big"
	"os"

	"github.com/fabricioricard/nkn-defi/internal/adapters/ethereum"
	"github.com/fabricioricard/nkn-defi/internal/adapters/nknclient"
)

type BridgeWithdrawWorker struct {
	nkn       *nknclient.Client
	eth       *ethereum.Client
	nknSender *nknclient.Wallet // carteira NKN com fundos
}

func NewBridgeWithdrawWorker(nkn *nknclient.Client, eth *ethereum.Client) *BridgeWithdrawWorker {
	return &BridgeWithdrawWorker{
		nkn: nkn,
		eth: eth,
		nknSender: nknclient.NewWallet(os.Getenv("BRIDGE_NKN_PRIVATE_KEY")),
	}
}

func (w *BridgeWithdrawWorker) Start(ctx context.Context) {
	w.eth.ListenBurnEvents(ctx, func(from common.Address, amount *big.Int, nknAddress string) {
		// Aplica taxa (ex: 0.1%)
		fee := new(big.Int).Div(amount, big.NewInt(1000)) // 0.1%
		sendAmount := new(big.Int).Sub(amount, fee)

		// Envia NKN mainnet para nknAddress
		txHash, err := w.nknSender.SendNKN(nknAddress, sendAmount)
		if err != nil {
			log.Printf("failed to send NKN withdrawal: %v", err)
			return
		}
		log.Printf("Sent %s NKN to %s (tx: %s)", sendAmount.String(), nknAddress, txHash)
		// Opcional: chamar função no contrato para marcar como completa
	})
}