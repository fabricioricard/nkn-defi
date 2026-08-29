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
	rpcAddrs []string
}

func NewClient(baseURL string) *Client {
	// Lista de nós RPC públicos (porta 30003) com fallback
	defaultAddrs := []string{
		"http://seed.nkn.org:30003",
		"http://mainnet-seed-0001.nkn.org:30003",
		"http://mainnet-seed-0002.nkn.org:30003",
		"http://mainnet-seed-0003.nkn.org:30003",
		"http://mainnet-seed-0004.nkn.org:30003",
	}

	// Se uma URL específica foi fornecida, coloca no início
	if baseURL != "" {
		addr := strings.TrimSpace(baseURL)
		addr = strings.TrimRight(addr, "/")
		if !strings.HasPrefix(addr, "http") {
			addr = "http://" + addr
		}
		defaultAddrs = append([]string{addr}, defaultAddrs...)
	}

	return &Client{rpcAddrs: defaultAddrs}
}

type Transaction struct {
	Hash   string
	Amount *big.Int
	From   string
	To     string
}

// GetRecentTransactions tenta cada nó RPC até obter sucesso.
func (c *Client) GetRecentTransactions(address string) ([]Transaction, error) {
	var lastErr error
	for _, rpcAddr := range c.rpcAddrs {
		txs, err := c.fetchFromRPC(rpcAddr, address)
		if err == nil {
			return txs, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("all RPC nodes failed: %v", lastErr)
}

func (c *Client) fetchFromRPC(rpcAddr, address string) ([]Transaction, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "getaddresstransactions",
		"params":  []interface{}{address},
		"id":      1,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := http.Post(rpcAddr, "application/json", bytes.NewReader(jsonData))
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

	// Aceita lista direta ou envelopada em "data"
	var txsRaw []struct {
		Hash   string `json:"hash"`
		Value  string `json:"value"`
		Data   string `json:"data"`
		Txid   string `json:"txid"`
		Amount string `json:"amount"`
		From   string `json:"from"`
		To     string `json:"to"`
	}

	// Tenta primeiro como array
	if err := json.Unmarshal(rpcResp.Result, &txsRaw); err != nil {
		// Tenta como objeto com campo "data"
		var envelop struct {
			Data []struct {
				Hash   string `json:"hash"`
				Value  string `json:"value"`
				Txid   string `json:"txid"`
				Amount string `json:"amount"`
				From   string `json:"from"`
				To     string `json:"to"`
			} `json:"data"`
		}
		if err2 := json.Unmarshal(rpcResp.Result, &envelop); err2 != nil {
			return nil, fmt.Errorf("parse transactions result: %w", err2)
		}
		for _, t := range envelop.Data {
			txsRaw = append(txsRaw, struct {
				Hash   string `json:"hash"`
				Value  string `json:"value"`
				Data   string `json:"data"`
				Txid   string `json:"txid"`
				Amount string `json:"amount"`
				From   string `json:"from"`
				To     string `json:"to"`
			}{t.Hash, t.Value, "", t.Txid, t.Amount, t.From, t.To})
		}
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
		hash := t.Hash
		if hash == "" {
			hash = t.Txid
		}
		txs = append(txs, Transaction{
			Hash:   hash,
			Amount: amt,
			From:   t.From,
			To:     t.To,
		})
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