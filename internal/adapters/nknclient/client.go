package nknclient

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/")}
}

type Transaction struct {
	Hash   string   `json:"hash"`
	Amount *big.Int `json:"amount"`
	Memo   string   `json:"memo"`
}

func (c *Client) GetRecentTransactions(address string) ([]Transaction, error) {
	url := fmt.Sprintf("%s/addresses/%s/transactions?limit=20", c.baseURL, address)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET transactions: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Data []struct {
			Txid    string `json:"txid"`
			Payload struct {
				Amount string `json:"amount"`
				Data   string `json:"data"`
			} `json:"payload"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse transactions: %w", err)
	}

	var txs []Transaction
	for _, d := range result.Data {
		amt, ok := new(big.Int).SetString(d.Payload.Amount, 10)
		if !ok {
			continue
		}
		memo := d.Payload.Data
		if decoded, err := hexDecode(memo); err == nil {
			memo = decoded
		}
		txs = append(txs, Transaction{
			Hash:   d.Txid,
			Amount: amt,
			Memo:   memo,
		})
	}
	return txs, nil
}

func hexDecode(s string) (string, error) {
	if len(s) < 2 || s[:2] != "0x" {
		return "", fmt.Errorf("not hex")
	}
	bytes, err := hex.DecodeString(s[2:])
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}