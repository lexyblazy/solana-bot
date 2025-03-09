package engine

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	// "fmt"
	// "io"
	"log"
	"os"

	"solana-bot/config"
	"solana-bot/db"
	"solana-bot/dexscreener"
	"solana-bot/helius"
	"solana-bot/jupiter"
	"solana-bot/wallet"

	"strings"
	"time"

	"github.com/leekchan/accounting"
)

type Engine struct {
	db     *db.SqlClient
	w      *wallet.WalletClient
	hs     *helius.Streamer
	hhc    *helius.HttpClient
	ds     *dexscreener.Client
	config *config.Config
	j      *jupiter.Client
	t      *Trader
}

func (e *Engine) DeleteProcessedLogs() {
	for {
		log.Println("DeleteProcessedLogs: Running")
		e.db.DeleteLogs()
		time.Sleep(30 * time.Second)
	}
}

func (e *Engine) ProcessLogs() {

	for {
		log.Println("ProcessLogs: Running")
		logs := e.db.GetUnProcessedLogs(100)

		logBatchSize := int(e.config.Engine.LogBatchSize)

		if len(logs) < logBatchSize {

			log.Printf("ProcessLogs: Skipping Until we have enough logs to process %d/%d \n", len(logs), logBatchSize)

			// wait for 1 minute before attempting to check for unprocessed logs again
			time.Sleep(1 * time.Minute)

			continue
		}

		var signatures []string

		for _, log := range logs {
			signatures = append(signatures, log.Signature)
		}

		txs, err := e.hhc.GetParsedTxs(signatures)

		if err != nil {

			log.Println(err)

			// move to the next iteration
			continue
		}

		for _, tx := range txs {

			for _, inc := range tx.Instructions {
				if inc.ProgramId == e.config.LiquidityPool.RaydiumProgramId {
					acct1, acct2 := inc.Accounts[8], inc.Accounts[9]
					newTokenAddress := ""

					if acct1 == e.config.Solana.NativeMint {
						newTokenAddress = acct2
					} else {
						newTokenAddress = acct1
					}

					e.db.InsertNewToken(newTokenAddress, tx.Signature)
				}
			}

		}

		// wait for 30 seconds for the next iteration
		time.Sleep(time.Second * 30)
	}

}

func (e *Engine) handleLogSubscribeMessage(message []byte) {

	var m helius.LogSubscribeMessage

	err := json.Unmarshal(message, &m)

	if err != nil {
		log.Printf("Failed to Unmarshal message  %s, %s\n", string(message), err)

		return

	}

	logs := m.Params.Result.Value.Logs

	for _, log := range logs {
		if strings.Contains(log, e.config.LiquidityPool.MigrationMessage) {
			e.db.InsertLog(m.Params.Result.Value.Signature)
		}
	}
}

func (e *Engine) RefreshTopTokensMetadata() {
	c := e.config.Engine.RefreshTopTokens
	batchSize := e.config.Engine.RefreshTokenMetadata.BatchSize

	for {
		log.Println("RefreshTopTokensMetadata... Running")
		tokens := e.db.GetTopTokens(c.MinMarketCap)

		if len(tokens) > 0 {
			log.Printf("Found %d top tokens ", len(tokens))

			for i := 0; i < len(tokens); i += batchSize {
				e.refreshTokensMetadata(tokens[i:min(i+batchSize, cap(tokens))])
				time.Sleep(1 * time.Second) // buffer to avoid being rate-limited by dexscreener
			}

		}

		time.Sleep(time.Duration(c.FrequencySeconds) * time.Second)
	}
}

func (e *Engine) refreshTokensMetadata(tokens []db.TokenEntity) {

	addresses := []string{}

	for _, token := range tokens {
		addresses = append(addresses, token.ContractAddress)
	}
	dexScreenerTokens := e.ds.GetTokenByAddress(addresses)

	if len(dexScreenerTokens) > 0 {
		e.db.UpdateTokenData(dexScreenerTokens)
	}

	e.db.UpdateTokensAsProcessed(addresses)
}

func (e *Engine) RefreshTokensMetadata() {

	refreshConfig := e.config.Engine.RefreshTokenMetadata
	minMarketCap := e.config.Engine.RefreshTopTokens.MinMarketCap

	for {
		log.Println("RefreshTokensMetadata: Running")

		tokens := e.db.GetTokensForProcessing(refreshConfig.BatchSize, refreshConfig.FrequencyMinutes, minMarketCap)

		if len(tokens) > 0 {
			log.Printf("RefreshTokensMetadata: Retrieved %d tokens from db for processing \n", len(tokens))

			e.refreshTokensMetadata(tokens)
		}

		time.Sleep(5 * time.Second)

	}
}

func (e *Engine) RemoveScamTokens() {

	scamTokensConfig := e.config.Engine.RemoveScamTokens

	for {

		log.Printf("RemoveScamTokens: Removing scam tokens where marketCap < $%v and older than %v hours \n",
			scamTokensConfig.MinMarketCap, scamTokensConfig.MinAgeHours)

		scamTokens := e.db.GetScamTokens(scamTokensConfig.MinMarketCap, scamTokensConfig.MinAgeHours)
		if len(scamTokens) > 0 {

			e.db.DeleteTokens(scamTokens)
		}

		time.Sleep(5 * time.Minute)
	}

}

func (e *Engine) CreateRugReport(t db.TokenEntity, m []db.MarketDataEntity) {

	ac := accounting.Accounting{Symbol: "$", Precision: 2}
	var mr []db.MarketDataEntity

	// want to limit number of market data records. Maximum of 5
	if len(m) > 6 {
		mr = append(mr, m[:3]...)
		mr = append(mr, m[len(m)-3:]...)
	} else {
		mr = append(mr, m...)
	}

	reportData := RugReportData{
		Token:      t,
		MarketData: mr,
	}

	var output bytes.Buffer

	report := `
symbol: ${{.Token.Symbol}}
C.A: {{.Token.ContractAddress}}
createdAt: {{formatDate .Token.PairCreatedAt}}

Timestamp MarketCap Liquidity
{{range .MarketData}}
{{formatTime .Timestamp}} {{formatMoney .MarketCap}} {{formatMoney .LiquidityUsd}}
{{end}}
	`

	tmpl, err := template.New(t.ContractAddress).Funcs(template.FuncMap{
		"formatTime": func(timestamp *time.Time) string {
			return timestamp.Format(time.TimeOnly)
		},

		"formatDate": func(timestamp *time.Time) string {
			return timestamp.Format(time.DateTime)
		},

		"formatMoney": func(val float64) string {
			return ac.FormatMoney(int64(val))
		},
	}).Parse(report)

	if err != nil {
		log.Println("CreateRugReport: Failed to init template", err)
		return
	}

	err = tmpl.Execute(&output, reportData)

	if err != nil {
		log.Println("CreateRugReport: Failed to execute template", err)
		return
	}

	fileName := fmt.Sprintf("reports/%s.txt", t.ContractAddress)
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		log.Println("CreateRugReport:", err)

		return
	}

	file.Write(output.Bytes())

}

func (e *Engine) GetRugsReport() {
	file, err := os.Open("./rug_tokens.txt")

	if err != nil {
		log.Println("Failed to open rug_tokens.txt file", err)

		return
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	addresses := []string{}

	for scanner.Scan() {
		address := scanner.Text()
		addresses = append(addresses, address)
	}

	tokens := e.db.GetTokensByContractAddress(addresses)

	for _, token := range tokens {
		marketData := e.db.GetTokenMarketData(token.ContractAddress)

		if len(marketData) == 0 {
			continue
		}

		e.CreateRugReport(token, marketData)

	}

}

func (e *Engine) Start() {

	// start trading engine
	go e.t.Start()
	// process RPC Logs
	go e.ProcessLogs()
	go e.DeleteProcessedLogs()

	// Delete scam tokens
	go e.RemoveScamTokens()

	// refresh token metadata
	go e.RefreshTopTokensMetadata()
	go e.RefreshTokensMetadata()

	// subscribe to helius websocket streaming
	go e.hs.SubscribeToLogs()
	go e.hs.ReadMessages()

	for msg := range e.hs.GetMessageChannel() {
		go e.handleLogSubscribeMessage(msg)
	}

}

func New(c *config.Config) *Engine {

	hhc := helius.NewHttpClient(&c.Helius)
	hs := helius.NewStreamer(&c.Helius)
	w := wallet.New(&c.Wallet, hhc)
	j := jupiter.New(&c.Jupiter)
	db := db.New(c.Engine.DSN)

	t := NewTrader(w, j, hhc, c, db)

	return &Engine{
		db:     db,
		hs:     hs,
		hhc:    hhc,
		config: c,
		ds:     dexscreener.New(&c.DexScreener),
		w:      w,
		j:      j,
		t:      t,
	}

}

func (s *Engine) Cleanup() {
	// close db, streamer
	s.db.Close()
	s.hs.Close()
}
