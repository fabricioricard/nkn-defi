package nknclient

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/nknorg/nkn-sdk-go"
)

type Wallet struct {
	mainAccount *nkn.Account
	mainWallet  *nkn.Wallet
	seed        []byte
}

// NewWallet cria uma carteira principal a partir da seed hexadecimal.
func NewWallet(privateKeyHex string) (*Wallet, error) {
	// Remove prefixo 0x se existir
	if len(privateKeyHex) >= 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}
	seed, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key hex: %w", err)
	}
	acc, err := nkn.NewAccount(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}
	wallet, err := nkn.NewWallet(acc, &nkn.WalletConfig{})
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}
	return &Wallet{
		mainAccount: acc,
		mainWallet:  wallet,
		seed:        seed,
	}, nil
}

// DeriveAddress gera um endereço determinístico a partir do índice.
// Usa SHA256(seed || index) como nova seed, criando uma conta temporária.
func (w *Wallet) DeriveAddress(index uint32) (string, error) {
	indexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(indexBytes, index)
	// Concatena seed + index
	combined := append(w.seed, indexBytes...)
	hash := sha256.Sum256(combined)

	// Cria uma conta derivada a partir do hash (nova seed)
	acc, err := nkn.NewAccount(hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to create derived account: %w", err)
	}
	return acc.WalletAddress(), nil
}

// GetDerivedWallet retorna uma carteira (com chave privada) para o índice derivado.
// Necessário para varrer fundos do endereço derivado.
func (w *Wallet) GetDerivedWallet(index uint32) (*nkn.Wallet, error) {
	indexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(indexBytes, index)
	combined := append(w.seed, indexBytes...)
	hash := sha256.Sum256(combined)

	acc, err := nkn.NewAccount(hash[:])
	if err != nil {
		return nil, err
	}
	return nkn.NewWallet(acc, &nkn.WalletConfig{})
}

// SendNKN envia NKN da carteira principal para um endereço.
func (w *Wallet) SendNKN(to string, amount *big.Int) (string, error) {
	txHash, err := w.mainWallet.Transfer(to, amount.String(), nil)
	if err != nil {
		return "", fmt.Errorf("transfer failed: %w", err)
	}
	return txHash, nil
}