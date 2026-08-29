package nknclient

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
)

type Client struct {
	rpcURL string
}

func NewClient(baseURL string) *Client {
	// Remove prefixos e barras finais para obter a URL RPC limpa
	url := strings.TrimPrefix(strings.TrimPrefix(baseURL, "http://"), "https://")
	url = strings.TrimRight(url, "/")
	return &Client{rpcURL: url}
}

type Transaction struct {
	Hash   string
	Amount *big.Int
	Memo   string
}

// GetRecentTransactions obtém as últimas transações de um endereço via JSON‑RPC.
func (c *Client) GetRecentTransactions(address string) ([]Transaction, error) {
	// Usa o endpoint RPC público da NKN
	rpcEndpoint := "https://" + c.rpcURL
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "getaddresstransactions",
		"params":  []string{address},
		"id":      1,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := http.Post(rpcEndpoint, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("POST RPC: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var rpcResp struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("parse RPC response: %w", err)
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
	}

	// O resultado pode ser uma lista de transações ou vazio
	var txsRaw []struct {
		Hash   string `json:"hash"`
		Value  string `json:"value"`
		Data   string `json:"data"`
		Txid   string `json:"txid"`
		Amount string `json:"amount"`
	}
	if err := json.Unmarshal(rpcResp.Result, &txsRaw); err != nil {
		return nil, fmt.Errorf("parse transactions result: %w", err)
	}

	var txs []Transaction
	for _, t := range txsRaw {
		amountStr := t.Value
		if amountStr == "" {
			amountStr = t.Amount
		}
		amt, ok := new(big.Int).SetString(amountStr, 10)
		if !ok {
			continue
		}
		tx := Transaction{
			Hash:   t.Hash,
			Amount: amt,
			Memo:   t.Data,
		}
		if tx.Hash == "" {
			tx.Hash = t.Txid
		}
		if decoded, err := hexDecode(tx.Memo); err == nil {
			tx.Memo = decoded
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func hexDecode(s string) (string, error) {
	if len(s) < 2 || s[:2] != "0x" {
		return "", fmt.Errorf("not hex")
	}
	b, err := hex.DecodeString(s[2:])
	if err != nil {
		return "", err
	}
	return string(b), nil
}