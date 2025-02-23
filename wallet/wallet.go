package wallet

import (
	"fmt"
	"log"
	"solana-bot/config"

	"github.com/gagliardetto/solana-go"
)

type WalletClient struct {
	privKey   solana.PrivateKey
	publicKey string
}

func New(c *config.WalletConfig) *WalletClient {

	pkey, err := solana.PrivateKeyFromBase58(c.PrivKey)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	if pkey.PublicKey().String() != c.Pubkey {
		log.Fatal("Public Key mismatch")
	}

	return &WalletClient{
		privKey:   pkey,
		publicKey: c.Pubkey,
	}
}

func (w *WalletClient) GetBalance() {

}
