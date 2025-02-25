package wallet

import (
	"log"
	"solana-bot/config"
	"solana-bot/helius"

	"github.com/gagliardetto/solana-go"
)

const (
	LAMPORT int = 1e9
)

type WalletClient struct {
	privKey   solana.PrivateKey
	publicKey string
	h         *helius.HttpClient
}

func New(c *config.WalletConfig, h *helius.HttpClient) *WalletClient {

	pkey, err := solana.PrivateKeyFromBase58(c.PrivKey)

	if err != nil {
		log.Println(err)
		return nil
	}

	if pkey.PublicKey().String() != c.Pubkey {
		log.Fatal("Public Key mismatch")
	}

	return &WalletClient{
		privKey:   pkey,
		publicKey: c.Pubkey,
		h:         h,
	}
}

func (w *WalletClient) GetBalance() float32 {
	balLamport := w.h.GetBalance(w.publicKey)

	balSol := float32(balLamport) / float32(LAMPORT)

	log.Printf("Bal sol: %f", balSol)

	return balSol

}
