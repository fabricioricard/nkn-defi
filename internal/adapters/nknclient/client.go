package nknclient

import (
	"bytes"
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
	url := strings.TrimSpace(baseURL)
	if url == "" {
		url = "https://mainnet-rpc-node-0001.nkn.org/mainnet/api/wallet"
	}
	return &Client{rpcURL: url}
}

// Transaction mantida para compatibilidade, mas não será usada.
type Transaction struct {
	Hash   string
	Amount *big.Int
	Memo   string
}

// GetAddressBalance consulta o saldo de um endereço NKN via JSON‑RPC.
func (c *Client) GetAddressBalance(address string) (*big.Int, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "getbalance",
		"params":  []string{address},
		"id":      1,
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

	// O resultado pode ser string ou número
	var balanceStr string
	if err := json.Unmarshal(rpcResp.Result, &balanceStr); err != nil {
		// tenta número
		var balanceNum json.Number
		if err2 := json.Unmarshal(rpcResp.Result, &balanceNum); err2 != nil {
			return nil, fmt.Errorf("parse balance result: %w", err2)
		}
		balanceStr = balanceNum.String()
	}

	balance, ok := new(big.Int).SetString(balanceStr, 10)
	if !ok {
		return nil, fmt.Errorf("invalid balance: %s", balanceStr)
	}
	return balance, nil
}