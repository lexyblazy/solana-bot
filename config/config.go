package config

type HeliusConfig struct {
	ApiKey       string `json:"apiKey"`
	RpcUrl       string `json:"rpcUrl"`
	WebSocketUrl string `json:"wsUrl"`
	BaseApiUrl   string `json:"baseApiUrl"`
}

type DexScreenerConfig struct {
	BaseUrl       string `json:"baseUrl"`
	SolanaChainId string `json:"solanaChainId"`
}

type Config struct {
	LiquidityPool struct {
		RaydiumProgramId string `json:"raydiumProgramId"`
		MigrationMessage string `json:"migrationMessage"`
	} `json:"liquidityPool"`

	Solana struct {
		NativeMint string `json:"nativeMint"`
	} `json:"solana"`

	Helius HeliusConfig `json:"helius"`

	Engine struct {
		DSN          string `json:"databaseName"`
		LogBatchSize int    `json:"processLogBatchSize"`

		RefreshTokenMetadata struct {
			BatchSize        int `json:"batchSize"`
			FrequencyMinutes int `json:"frequencyMinutes"`
		} `json:"refreshTokenMetadata"`

		RemoveScamTokens struct {
			MinMarketCap int `json:"minMarketCap"`
			MinAgeHours   int `json:"minAgeHours"`
		} `json:"removeScamTokens"`
	} `json:"engine"`

	DexScreener DexScreenerConfig `json:"dexscreener"`
}
