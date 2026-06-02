package nknclient

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"

	"github.com/nknorg/nkn-sdk-go"
)

// Client para consultas à API REST da NKN.
type Client struct {
	baseURL string
}

// Transaction representa uma transação retornada pela API.
type Transaction struct {
	Hash   string   `json:"hash"`
	Amount *big.Int `json:"amount"`
	Memo   string   `json:"memo"` // campo de dados (hex ou texto)
}

// NewClient cria um cliente REST. baseURL ex: "https://openapi.nkn.org/v1"
func NewClient(baseURL string) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/")}
}

// GetRecentTransactions obtém as últimas transações de um endereço.
func (c *Client) GetRecentTransactions(address string) ([]Transaction, error) {
	url := fmt.Sprintf("%s/addresses/%s/transactions?limit=20", c.baseURL, address)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET transactions: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	// A resposta da API NKN varia. Vamos assumir um envelope:
	// {"data": [ {"txid":"...", "attributes": {...}, "payload": {"amount":"...", "data":"..."} } ]}
	var result struct {
		Data []struct {
			Txid string `json:"txid"`
			Payload struct {
				Amount string `json:"amount"`
				Data   string `json:"data"` // memo em hex
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
		// Decodifica o memo (hexadecimal)
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

// Wallet agrupa uma conta NKN para assinar e enviar transações.
type Wallet struct {
	account *nkn.Account
}

// NewWallet cria uma carteira a partir de uma chave privada (formato WIF ou hex).
func NewWallet(privateKey string) (*Wallet, error) {
	acc, err := nkn.NewAccount(privateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}
	return &Wallet{account: acc}, nil
}

// SendNKN envia uma quantidade de NKN (em unidades inteiras) para um endereço.
func (w *Wallet) SendNKN(to string, amount *big.Int) (string, error) {
	txHash, err := w.account.Transfer(to, amount.String(), nil)
	if err != nil {
		return "", fmt.Errorf("transfer: %w", err)
	}
	return txHash, nil
}

func hexDecode(s string) (string, error) {
	if len(s) < 2 || s[:2] != "0x" {
		return "", fmt.Errorf("not hex")
	}
	bytes, err := hex.DecodeString(s[2:]) // precisa import "encoding/hex"
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}