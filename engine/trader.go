package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"solana-bot/config"
	"solana-bot/db"
	"solana-bot/helius"
	"solana-bot/jupiter"
	"solana-bot/utils"
	"solana-bot/wallet"
	"sync"
	"time"
)

type Trader struct {
	c  *config.Config
	db *db.SqlClient
	h  *helius.HttpClient
	j  *jupiter.Client
	w  *wallet.WalletClient

	cache map[uint64]bool
	mu    sync.RWMutex
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

// A Buy is swapping native sol to the "meme" token address, a SwapFromNativeSol
func (t *Trader) buyToken(mintAddress string, amountSol float32) (string, error) {

	exponential := float32(math.Pow(10, float64(t.getTokenDecimals(mintAddress))))
	amountLamport := int(amountSol * exponential)

	bal := t.w.GetBalance()

	if bal < amountLamport {

		errMessage := fmt.Sprintf("buyToken: Insufficient Balance, Expected >= %d, Got = %d \n", amountLamport, bal)
		log.Print(errMessage)

		return "", errors.New(errMessage)
	}

	return t.swap(SwapTokenParams{
		InputMint:  t.c.Solana.NativeMint,
		OutputMint: mintAddress,
		Amount:     amountLamport,
	})

}

func (t *Trader) isLocked(id uint64) bool {
	// obtain a reader's lock
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, ok := t.cache[id]

	return ok
}

func (t *Trader) acquireLock(id uint64) bool {
	if t.isLocked(id) {
		return false
	}

	// obtain a writer's lock
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cache[id] = true

	return true
}

func (t *Trader) releaseLock(id uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.cache, id)
}

// A Sell is swapping the "meme" token address to native sol, a SwapToNativeSol
func (t *Trader) sellToken(mintAddress string, rules *string) (string, error) {

	bal := t.w.GetTokenBalance(mintAddress)
	log.Println("TokenBalance", bal)

	if rules == nil {
		// sell everything
		return t.swap(SwapTokenParams{
			InputMint:  mintAddress,
			OutputMint: t.c.Solana.NativeMint,
			Amount:     bal,
		})

	}

	// TODO - implement way to sell parts of the token
	// swapRules := utils.Deserialize[db.SwapRules](*rules)
	// exponential := float32(math.Pow(10, float64(t.getTokenDecimals(mintAddress))))
	// atomicUnit := int(amountToken * exponential)
	atomicUnit := bal

	if bal < atomicUnit {

		errMessage := fmt.Sprintf("SellToken: Insufficient Balance, Expected >= %d, Got = %d \n", atomicUnit, bal)
		log.Println(errMessage)

		return "", errors.New(errMessage)
	}

	return t.swap(SwapTokenParams{
		InputMint:  mintAddress,
		OutputMint: t.c.Solana.NativeMint,
		Amount:     atomicUnit,
	})

}

func (t *Trader) swap(params SwapTokenParams) (string, error) {

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

func (t *Trader) processPendingTrades() {

	for {
		trades := t.db.GetPendingTrades()

		if len(trades) > 0 {
			fmt.Printf("ProcessPendingTrades: Found %v pending trades \n", len(trades))
			for _, tr := range trades {
				go t.executeTrade(tr)
			}
		}

		time.Sleep(1 * time.Minute)
	}
}

func (t *Trader) executeTrade(tr db.SwapTradeEntity) {

	// acquire the lock
	lock := t.acquireLock(tr.Id)

	if !lock {
		fmt.Printf("Id = %d is already locked for processing \n", tr.Id)

		return
	}

	var txHash string

	// for buy orders we set the amount
	if tr.AmountDetails != nil {

		amtDetails := utils.Deserialize[db.AmountDetails](*tr.AmountDetails)

		hash, err := t.buyToken(tr.ToToken, amtDetails.QuantitySol)

		if err != nil {
			log.Println("executeTrade: buy failed", err)
		}

		txHash = hash

	} else {

		hash, err := t.sellToken(tr.FromToken, tr.Rules)

		if err != nil {
			log.Println("executeTrade: sell failed", err)
		}

		txHash = hash

	}

	t.db.UpdateSwapOrder(txHash, tr.Id)

	// release lock after processing
	t.releaseLock(tr.Id)

}

func (t *Trader) loadTrades() {
	file, err := os.Open("./trades.json")

	if err != nil {
		log.Println("Failed open trades.json file")
		return
	}

	var swapTrades []db.SwapTradeEntity

	json.NewDecoder(file).Decode(&swapTrades)

	for _, st := range swapTrades {
		t.db.InsertSwapOrder(st)
	}

}

func (t *Trader) Start() {
	// t.loadTrades()
	t.processPendingTrades()
}

func NewTrader(w *wallet.WalletClient, j *jupiter.Client, h *helius.HttpClient, c *config.Config, db *db.SqlClient) *Trader {
	return &Trader{
		w:     w,
		j:     j,
		h:     h,
		c:     c,
		db:    db,
		cache: make(map[uint64]bool),
	}
}
