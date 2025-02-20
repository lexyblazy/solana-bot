package dexscreener

type TokensByAddress struct {
	ChainID     string `json:"chainId"`
	DexID       string `json:"dexId"`
	URL         string `json:"url"`
	PairAddress string `json:"pairAddress"`
	BaseToken   struct {
		Address string `json:"address"`
		Name    string `json:"name"`
		Symbol  string `json:"symbol"`
	} `json:"baseToken"`
	QuoteToken struct {
		Address string `json:"address"`
		Name    string `json:"name"`
		Symbol  string `json:"symbol"`
	} `json:"quoteToken"`
	PriceNative string `json:"priceNative"`
	PriceUsd    string `json:"priceUsd"`
	Txns        struct {
		M5 struct {
			Buys  int `json:"buys"`
			Sells int `json:"sells"`
		} `json:"m5"`
		H1 struct {
			Buys  int `json:"buys"`
			Sells int `json:"sells"`
		} `json:"h1"`
		H6 struct {
			Buys  int `json:"buys"`
			Sells int `json:"sells"`
		} `json:"h6"`
		H24 struct {
			Buys  int `json:"buys"`
			Sells int `json:"sells"`
		} `json:"h24"`
	} `json:"txns"`
	Volume struct {
		H24 float64 `json:"h24"`
		H6  int     `json:"h6"`
		H1  int     `json:"h1"`
		M5  int     `json:"m5"`
	} `json:"volume"`
	PriceChange struct {
		H24 int `json:"h24"`
	} `json:"priceChange"`
	Liquidity struct {
		Usd   float64 `json:"usd"`
		Base  int     `json:"base"`
		Quote float64 `json:"quote"`
	} `json:"liquidity"`
	Fdv           int   `json:"fdv"`
	MarketCap     int   `json:"marketCap"`
	PairCreatedAt int64 `json:"pairCreatedAt"`
}
