package dexscreener

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"solana-bot/config"
	"strings"
)

type Client struct {
	config *config.DexScreenerConfig
}

func getTokensByAddress() {
	// rate-limit 300 requests per minute
}

func (c *Client) GetTokenByAddress(addresses []string) []TokensByAddress {

	url := fmt.Sprintf("%s/tokens/v1/%s/%s", c.config.BaseUrl, c.config.SolanaChainId, strings.Join(addresses, ","))
	resp, err := http.Get(url)

	if err != nil {
		log.Println("GetTokenByAddress:", err)

		return nil
	}
	defer resp.Body.Close()

	var tokens []TokensByAddress

	json.NewDecoder(resp.Body).Decode(&tokens)

	return tokens
}

func New(config *config.DexScreenerConfig) *Client {
	return &Client{config}
}
