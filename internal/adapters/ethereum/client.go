package ethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	eth      *ethclient.Client
	minterPK *ecdsa.PrivateKey
	minterAddr common.Address
	chainID  *big.Int
	wNKN     *WrappedNKN
}

func NewClient(rpcURL string) (*Client, error) {
	eth, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}
	chainID, err := eth.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	privateKeyHex := os.Getenv("MINTER_PRIVATE_KEY")
	if privateKeyHex == "" {
		return nil, fmt.Errorf("MINTER_PRIVATE_KEY env var not set")
	}
	// Remove prefixo 0x se existir
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	minterAddr := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &Client{
		eth:      eth,
		minterPK: privateKey,
		minterAddr: minterAddr,
		chainID:  chainID,
	}, nil
}

func (c *Client) LoadWNKNContract(address string) error {
	contractAddr := common.HexToAddress(address)
	var err error
	c.wNKN, err = NewWrappedNKN(contractAddr, c.eth)
	return err
}

func (c *Client) BlockNumber(ctx context.Context) (uint64, error) {
	return c.eth.BlockNumber(ctx)
}