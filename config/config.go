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

type WalletConfig struct {
	Pubkey  string `json:"publicKey"`
	PrivKey string `json:"privateKey"`
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
			BatchSize        int `json:"batchSize"` // this is a dexscreener limitation
			FrequencyMinutes int `json:"frequencyMinutes"`
		} `json:"refreshTokenMetadata"`

		RemoveScamTokens struct {
			MinMarketCap int `json:"minMarketCap"`
			MinAgeHours  int `json:"minAgeHours"`
		} `json:"removeScamTokens"`

		RefreshTopTokens struct {
			MinMarketCap     int `json:"minMarketCap"`
			FrequencySeconds int `json:"frequencySeconds"`
		} `json:"refreshTopTokens"`
	} `json:"engine"`

	DexScreener DexScreenerConfig `json:"dexscreener"`

	Wallet WalletConfig `json:"wallet"`
}
