package wallet

import (
	"log"
	"solana-bot/config"
	"solana-bot/helius"
	"strconv"

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

func (w *WalletClient) GetTokenBalance(mint string) int {

	result := w.h.GetTokenAccountsByOwner(w.publicKey, mint)

	if result == nil {
		return 0
	}

	amount := 0

	for _, v := range result.Result.Value {
		if v.Account.Data.Parsed.Info.Mint == mint {
			val, err := strconv.Atoi(v.Account.Data.Parsed.Info.TokenAmount.Amount)

			if err != nil {
				log.Println("Err converting tokenAmount to integer", err)
			} else {
				amount = val
			}
		}
	}

	return amount

}
