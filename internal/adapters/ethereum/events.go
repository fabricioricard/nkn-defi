package ethereum

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

// FilterBurnEvents busca eventos BurnRequested emitidos a partir do bloco fromBlock.
func (c *Client) FilterBurnEvents(ctx context.Context, fromBlock int64) ([]*WrappedNKNBurnRequested, error) {
	opts := bind.FilterOpts{
		Start:   uint64(fromBlock),
		End:     nil, // até o último bloco
		Context: ctx,
	}
	// Assinatura: FilterBurnRequested(opts *bind.FilterOpts, from []common.Address) (Iterator, error)
	iterator, err := c.wNKN.FilterBurnRequested(&opts, nil)
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var events []*WrappedNKNBurnRequested
	for iterator.Next() {
		events = append(events, iterator.Event)
	}
	return events, iterator.Error()
}