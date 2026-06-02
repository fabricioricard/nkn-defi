package ethereum

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func (c *Client) MintWNKN(to common.Address, amount *big.Int, mainnetTxHash [32]byte) (string, error) {
	auth, err := bind.NewKeyedTransactorWithChainID(c.minterPK, c.chainID)
	if err != nil {
		return "", err
	}
	tx, err := c.wNKN.Mint(auth, to, amount, mainnetTxHash)
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), nil
}