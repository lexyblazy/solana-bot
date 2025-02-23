package engine

import "solana-bot/db"

type RugReportData struct {
	Token db.TokenEntity
	MarketData []db.MarketDataEntity
}
