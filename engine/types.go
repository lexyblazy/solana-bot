package engine

import "solana-bot/db"

type RugReportData struct {
	Token      db.TokenEntity
	MarketData []db.MarketDataEntity
}

type Trades struct {
	Buy  *db.SwapTradeEntity `json:"buy"`
	Sell *db.SwapTradeEntity `json:"sell"`
}
