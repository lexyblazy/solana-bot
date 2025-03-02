package wallet

import (
	"log"
	"solana-bot/config"
	"solana-bot/helius"
	"strconv"

	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
)

const (
	LAMPORT int = 1e9
)

type WalletClient struct {
	privKey   solana.PrivateKey
	PublicKey string
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
		PublicKey: c.Pubkey,
		h:         h,
	}
}

func (w *WalletClient) GetBalance() int {
	balLamport := w.h.GetBalance(w.PublicKey)

	balSol := float32(balLamport) / float32(LAMPORT)

	log.Printf("Bal readable sol: %f", balSol)

	return balLamport

}

func (w *WalletClient) GetTokenBalance(mint string) int {

	result := w.h.GetTokenAccountsByOwner(w.PublicKey, mint)

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

func (w *WalletClient) CreateSignedTxMessage(message string) string {

	tx, err := solana.TransactionFromBase64(message)

	if err != nil {
		log.Println("CreateTx: TransactionFromBase64", err)

		return ""
	}

	tx.Sign(func(p solana.PublicKey) *solana.PrivateKey {
		return &w.privKey
	})

	txBytes, err := tx.MarshalBinary()

	if err != nil {
		log.Println("CreateTx: MarshalBinary", err)

		return ""
	}

	return base58.Encode(txBytes)

}
