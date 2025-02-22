package engine

import (
	"encoding/json"
	"log"

	"solana-bot/config"
	"solana-bot/db"
	"solana-bot/dexscreener"
	"solana-bot/helius"

	"strings"
	"time"
)

type Engine struct {
	db *db.SqlClient

	hs  *helius.Streamer
	hhc *helius.HttpClient

	ds *dexscreener.Client

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
		log.Println("Failed to Unmarshal message")

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

func New(c *config.Config) *Engine {

	return &Engine{
		db:     db.New(c.Engine.DSN),
		hs:     helius.NewStreamer(&c.Helius),
		hhc:    helius.NewHttpClient(&c.Helius),
		config: c,
		ds:     dexscreener.New(&c.DexScreener),
	}

}

func (e *Engine) Start() {

	go e.ProcessLogs()
	go e.DeleteProcessedLogs()
	go e.RefreshTokensMetadata()

	go e.hs.SubscribeToLogs()
	go e.hs.ReadMessages()

	for msg := range e.hs.GetMessageChannel() {
		go e.handleLogSubscribeMessage(msg)
	}

}

func (s *Engine) Cleanup() {
	// close db, streamer
	s.db.Close()
	s.hs.Close()
}
