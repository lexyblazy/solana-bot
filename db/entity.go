package db

import (
	"time"
)

type RpcLog struct {
	Id          uint64
	Signature   string
	CreatedAt   time.Time
	ProcessedAt *time.Time // nullable field
}

type Token struct {
	Id              uint64
	ContractAddress string
	CreatedAt       time.Time
	LastProcessedAt *time.Time // nullable field
	Symbol          *string    // nullable field
	MarketCap       *float64   // nullable filed
	PairCreatedAt   *time.Time // nullable filed
}
