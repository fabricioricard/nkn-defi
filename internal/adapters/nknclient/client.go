package nknclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
)

type Client struct {
	rpcURL string
}

func NewClient(baseURL string) *Client {
	url := strings.TrimSpace(baseURL)
	if url == "" {
		url = "https://explorer.nkn.org/api"
	}
	return &Client{rpcURL: url}
}

// GetAddressBalance consulta o saldo de um endereço usando o método nativo getbalancebyaddr.
// Suporta múltiplos formatos de resposta (string, número, objeto).
func (c *Client) GetAddressBalance(address string) (*big.Int, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "getbalancebyaddr",
		"params": map[string]interface{}{
			"address": address,
		},
		"id": 1,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := http.Post(c.rpcURL, "application/json", bytes.NewReader(jsonData))
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

	// Tenta diferentes formatos

	// 1. String simples
	var balanceStr string
	if err := json.Unmarshal(rpcResp.Result, &balanceStr); err == nil {
		if balance, ok := new(big.Int).SetString(balanceStr, 10); ok {
			return balance, nil
		}
	}

	// 2. Número
	var balanceNum json.Number
	if err := json.Unmarshal(rpcResp.Result, &balanceNum); err == nil {
		balanceStr = balanceNum.String()
		if balance, ok := new(big.Int).SetString(balanceStr, 10); ok {
			return balance, nil
		}
	}

	// 3. Objeto com chaves comuns
	var obj map[string]interface{}
	if err := json.Unmarshal(rpcResp.Result, &obj); err == nil {
		for _, key := range []string{"amount", "balance", "value", "Balance", "Amount", "Result"} {
			if v, exists := obj[key]; exists {
				switch val := v.(type) {
				case string:
					balanceStr = val
				case float64:
					balanceStr = strconv.FormatFloat(val, 'f', 0, 64)
				case json.Number:
					balanceStr = val.String()
				}
				if balanceStr != "" {
					if balance, ok := new(big.Int).SetString(balanceStr, 10); ok {
						return balance, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("unable to extract balance from result")
}

// GetRecentTransactions mantida para compatibilidade, mas retorna erro (não suportado por nós RPC).
func (c *Client) GetRecentTransactions(address string) ([]Transaction, error) {
	return nil, fmt.Errorf("not implemented: use balance polling")
}

type Transaction struct {
	Hash   string
	Amount *big.Int
	From   string
	To     string
}