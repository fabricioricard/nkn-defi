package nknclient

import (
	"errors"
	"math/big"
)

type Client struct{}

func NewClient(baseURL string) *Client {
	return &Client{}
}

type Transaction struct {
	Hash   string
	Amount *big.Int
	Memo   string
}

func (c *Client) GetRecentTransactions(address string) ([]Transaction, error) {
	// TODO: implementar chamada real à API NKN
	return nil, errors.New("not implemented")
}

type Wallet struct{}

func NewWallet(privateKey string) (*Wallet, error) {
	// TODO: implementar carteira NKN real
	return &Wallet{}, nil
}

func (w *Wallet) SendNKN(to string, amount *big.Int) (string, error) {
	// TODO: implementar envio de transação NKN
	return "", errors.New("not implemented")
}