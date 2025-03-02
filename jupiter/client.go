package jupiter

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"solana-bot/config"
)

type Client struct {
	config *config.JupiterConfig
}

func (c *Client) GetQuote(params GetQuoteParams) *GetQuoteResponse {

	url := fmt.Sprintf("%s/swap/v1/quote?%s", c.config.BaseUrl, params.toQueryString())

	var quote GetQuoteResponse

	resp, err := http.Get(url)

	if err != nil {
		log.Println("GetQuote: ", err)
		return nil
	}

	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&quote)

	return &quote

}

func New(c *config.JupiterConfig) *Client {
	return &Client{config: c}
}
