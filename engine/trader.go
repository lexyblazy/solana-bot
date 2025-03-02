package engine

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"solana-bot/config"
	"solana-bot/db"
	"solana-bot/helius"
	"solana-bot/jupiter"
	"solana-bot/wallet"
)

type Trader struct {
	c  *config.Config
	db *db.SqlClient
	h  *helius.HttpClient
	j  *jupiter.Client
	w  *wallet.WalletClient
}

type SwapTokenParams struct {
	InputMint  string
	OutputMint string
	Amount     int
}

func (p SwapTokenParams) ToString() string {
	return fmt.Sprintf("fromToken = %s, toToken = %s, amount =%d", p.InputMint, p.OutputMint, p.Amount)
}

func (t *Trader) getTokenDecimals(mintAddress string) int {
	if mintAddress == t.c.Solana.UsdcMint || mintAddress == t.c.Solana.UsdtMint {
		return 6
	}

	return 9
}

func (t *Trader) BuyToken(mintAddress string, amountSol float32) {

	exponential := float32(math.Pow(10, float64(t.getTokenDecimals(mintAddress))))
	amountLamport := int(amountSol * exponential)

	bal := t.w.GetBalance()

	if bal < amountLamport {

		log.Printf("BuyToken: Insufficient Balance, Expected >= %d, Got = %d \n", amountLamport, bal)

		return
	}

	t.Swap(SwapTokenParams{
		InputMint:  t.c.Solana.NativeMint,
		OutputMint: mintAddress,
		Amount:     amountLamport,
	})
}

func (t *Trader) SellToken(mintAddress string, amountToken float32) {

	exponential := float32(math.Pow(10, float64(t.getTokenDecimals(mintAddress))))
	atomicUnit := int(amountToken * exponential)
	bal := t.w.GetTokenBalance(mintAddress)

	log.Println("TokenBalance", bal)

	if bal < atomicUnit {

		log.Printf("SellToken: Insufficient Balance, Expected >= %d, Got = %d \n", atomicUnit, bal)

		return
	}

	t.Swap(SwapTokenParams{
		InputMint:  mintAddress,
		OutputMint: t.c.Solana.NativeMint,
		Amount:     atomicUnit,
	})

}

func (t *Trader) Swap(params SwapTokenParams) (string, error) {

	quote := t.j.GetQuote(jupiter.GetQuoteParams{
		InputMint:   params.InputMint,
		OutputMint:  params.OutputMint,
		Amount:      float64(params.Amount),
		SlippageBps: t.c.Jupiter.SlippageBps,
	})

	if quote == nil || len(quote.RoutePlan) < 1 {

		return "", fmt.Errorf("no quote found for swap: %s", params.ToString())
	}

	swapTx := t.j.BuildSwapTransaction(quote, t.w.PublicKey)

	if swapTx == nil {
		return "", fmt.Errorf("failed to BuildSwapTransaction: %s", params.ToString())
	}

	signedMessage := t.w.CreateSignedTxMessage(swapTx.SwapTransaction)
	txHash := t.h.SendTransaction(signedMessage)

	log.Println("Swap Completed... txHash", txHash)

	return txHash, nil

}

func (t *Trader) Demo() {
	t.BuyToken(t.c.Solana.UsdcMint, 20)
	t.SellToken(t.c.Solana.UsdcMint, 0.000057)
}

func (t *Trader) LoadTrades() {
	file, err := os.Open("./trades.json")

	if err != nil {
		log.Println("Failed open trades.json file")
		return
	}

	var swapTrades []Trades

	json.NewDecoder(file).Decode(&swapTrades)

	for _, st := range swapTrades {
		if st.Buy != nil {

			t.db.InsertBuyOrder(*st.Buy)
		}
		
		if st.Sell != nil {
			t.db.InsertSellOrder(*st.Sell)
		}
	}

}

func (t *Trader) Start() {
	t.LoadTrades()
}

func NewTrader(w *wallet.WalletClient, j *jupiter.Client, h *helius.HttpClient, c *config.Config, db *db.SqlClient) *Trader {
	return &Trader{w: w, j: j, h: h, c: c, db: db}
}
