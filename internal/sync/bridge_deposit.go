package sync

import (
	"context"
	"log"
	"math/big"
	"os"

	"github.com/fabricioricard/nkn-defi/internal/adapters/ethereum"
	"github.com/fabricioricard/nkn-defi/internal/adapters/nknclient"
)

type BridgeDepositWorker struct {
	nkn     *nknclient.Client
	eth     *ethereum.Client
	depositAddr string // endereço NKN que recebe depósitos
}

func NewBridgeDepositWorker(nkn *nknclient.Client, eth *ethereum.Client) *BridgeDepositWorker {
	return &BridgeDepositWorker{
		nkn:     nkn,
		eth:     eth,
		depositAddr: os.Getenv("BRIDGE_NKN_ADDRESS"),
	}
}

func (w *BridgeDepositWorker) Start(ctx context.Context) {
	// Polling a cada 15 segundos
	// Em produção, usar websocket ou subscription se disponível
	for {
		select {
		case <-ctx.Done():
			return
		default:
			txs, err := w.nkn.GetRecentTransactions(w.depositAddr)
			if err != nil {
				log.Printf("error fetching NKN txs: %v", err)
				continue
			}
			for _, tx := range txs {
				if tx.Value.Cmp(big.NewInt(0)) > 0 {
					// Determinar endereço Ethereum associado a essa transação
					// (precisa de um mapa prévio, por exemplo via API)
					ethAddr := getUserETHAddressForDeposit(tx.ID) // implementar mapeamento
					if ethAddr != "" {
						txHash := convertToBytes32(tx.Hash)
						w.eth.MintWNKN(common.HexToAddress(ethAddr), tx.Value, txHash)
						log.Printf("Minted %s wNKN to %s", tx.Value.String(), ethAddr)
						// Marcar transação como processada no banco
					}
				}
			}
			// sleep 15s
			time.Sleep(15 * time.Second)
		}
	}
}