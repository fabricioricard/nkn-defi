package ethereum

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func (c *Client) MintWNKN(to common.Address, amount *big.Int, mainnetTxHash [32]byte) (string, error) {
	tx, err := c.wNKN.Mint(c.buildTransactOpts(), to, amount, mainnetTxHash)
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}

func (c *Client) buildTransactOpts() *bind.TransactOpts {
	// utiliza a chave privada e chainID para assinar
	// implementação padrão: bind.NewKeyedTransactorWithChainID
	return bind.NewKeyedTransactorWithChainID(c.minterPK, c.chainID)
}