package nknclient

import (
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
	// Mantém a URL exata, removendo apenas barra final
	url := strings.TrimRight(baseURL, "/")
	if url == "" {
		url = "https://mainnet-rpc-node-0001.nkn.org/mainnet/api/wallet"
	}
	return &Client{baseURL: url}
}

type Transaction struct {
	Hash   string
	Amount *big.Int
	From   string
	To     string
}

// GetRecentTransactions obtém as últimas transações de um endereço via API REST (GET).
func (c *Client) GetRecentTransactions(address string) ([]Transaction, error) {
	url := fmt.Sprintf("%s/transactions?address=%s&limit=20", c.baseURL, address)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET transactions: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	// Estrutura da resposta real (conforme você observou)
	var apiResp struct {
		Data struct {
			List []struct {
				Hash      string `json:"Hash"`
				FromAddr  string `json:"FromAddr"`
				ToAddr    string `json:"ToAddr"`
				Value     string `json:"Value"`
				Fee       string `json:"Fee"`
				Timestamp string `json:"Timestamp"`
			} `json:"list"`
		} `json:"Data"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("parse transactions: %w", err)
	}

	var txs []Transaction
	for _, t := range apiResp.Data.List {
		amt, ok := new(big.Int).SetString(t.Value, 10)
		if !ok {
			continue
		}
		txs = append(txs, Transaction{
			Hash:   t.Hash,
			Amount: amt,
			From:   t.FromAddr,
			To:     t.ToAddr,
		})
	}
	return txs, nil
}