package ethereum

import (
	"context"
)

func (c *Client) ListenBurnEvents(ctx context.Context) (<-chan *WrappedNKNBurnRequested, error) {
	ch := make(chan *WrappedNKNBurnRequested)
	sub, err := c.wNKN.WatchBurnRequested(
		nil, // opts (nil = default)
		ch,  // sink
		nil, // from []common.Address (nil = any)
	)
	if err != nil {
		return nil, err
	}
	go func() {
		defer sub.Unsubscribe()
		<-ctx.Done()
	}()
	return ch, nil
}