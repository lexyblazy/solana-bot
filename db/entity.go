package db

import (
	"time"
)

type RpcLogEntity struct {
	Id          uint64
	Signature   string
	CreatedAt   time.Time
	ProcessedAt *time.Time // nullable field
}

type TokenEntity struct {
	Id              uint64
	ContractAddress string
	CreatedAt       time.Time
	LastProcessedAt *time.Time // nullable field
	Symbol          *string    // nullable field
	MarketCap       *float64   // nullable filed
	PairCreatedAt   *time.Time // nullable filed
}

type MarketDataEntity struct {
	Id              uint64
	Timestamp       time.Time
	MarketCap       float64
	Fdv             float64
	LiquidityUsd    float64
	PriceNative     float64
	PriceUsd        float64
	ContractAddress string
}

type SwapRules struct {
	TakeProfit float32 `json:"takeProfit"`
	StopLoss   float32 `json:"stopLoss"`
}

type AmountDetails struct {
	QuantitySol float32 `json:"quantitySol"`
}

type SwapTradeEntity struct {
	Id              uint64 `json:"id"`
	CreatedAt       time.Time
	LastProcessedAt *time.Time // nullable field
	ExecutedAt      *time.Time // nullable field

	FromToken string `json:"fromToken"`
	ToToken   string `json:"toToken"`

	TxHash *string    // nullable field
	Rules  *SwapRules `json:"rules"` // nullable field

	AmountDetails *AmountDetails `json:"amountDetails"` // nullable field
}
