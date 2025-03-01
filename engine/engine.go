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
}

func (e *Engine) DeleteProcessedLogs() {
	for {
		e.db.DeleteLogs()
		time.Sleep(30 * time.Second)
	}
}

func (e *Engine) ProcessLogs() {

	for {
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

func (e *Engine) RefreshTokensMetadata() {

	refreshConfig := e.config.Engine.RefreshTokenMetadata

	for {
		addresses := []string{}
		tokens := e.db.GetTokensForProcessing(refreshConfig.BatchSize, refreshConfig.FrequencyMinutes)

		if len(tokens) > 0 {
			log.Printf("RefreshTokensMetadata: Retrieved %d tokens from db for processing \n", len(tokens))

			for _, token := range tokens {
				addresses = append(addresses, token.ContractAddress)
			}

			dexScreenerTokens := e.ds.GetTokenByAddress(addresses)
			log.Printf("RefreshTokensMetadata: Retrieved %d tokens from dexscreener \n", len(dexScreenerTokens))

			if len(dexScreenerTokens) > 0 {
				e.db.UpdateTokenData(dexScreenerTokens)
				log.Printf("RefreshTokensMetadata: Updated token metadata for %d tokens \n", len(dexScreenerTokens))
			}

			e.db.UpdateTokensAsProcessed(addresses)
		} else {
			log.Printf("RefreshTokensMetadata: is configured for every  %d minutes. You can adjust the schedule \n", refreshConfig.FrequencyMinutes)
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

	go e.RemoveScamTokens()
	go e.ProcessLogs()
	go e.DeleteProcessedLogs()
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

	return &Engine{
		db:     db.New(c.Engine.DSN),
		hs:     helius.NewStreamer(&c.Helius),
		hhc:    hhc,
		config: c,
		ds:     dexscreener.New(&c.DexScreener),
		w:      wallet.New(&c.Wallet, hhc),
	}

}

func (s *Engine) Cleanup() {
	// close db, streamer
	s.db.Close()
	s.hs.Close()
}
