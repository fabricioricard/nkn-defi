package ethereum

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
)

func (c *Client) ListenBurnEvents(ctx context.Context, handler func(from common.Address, amount *big.Int, nknAddress string)) error {
	ch := make(chan *WrappedNKNBurnRequested)
	sub, err := c.wNKN.WatchBurnRequested(nil, ch, nil, nil)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-sub.Err():
			log.Printf("subscription error: %v", err)
			return err
		case event := <-ch:
			handler(event.From, event.Amount, event.NknMainnetAddress)
		}
	}
}