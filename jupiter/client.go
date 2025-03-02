package jupiter

import (
	"bytes"
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

func (c *Client) BuildSwapTransaction(quote *GetQuoteResponse, publicKey string) *BuildSwapTransactionResponseBody {

	body, err := json.Marshal(BuildSwapTransactionRequestBody{
		QuoteResponse:           *quote,
		UserPublicKey:           publicKey,
		DynamicComputeUnitLimit: true,
		DynamicSlippage:         true,
	})

	if err != nil {
		log.Println("failed to json.Marshal body", err)
		return nil
	}

	url := fmt.Sprintf("%s/swap/v1/swap", c.config.BaseUrl)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))

	if err != nil {
		log.Println("BuildSwapTransaction: Api call failed", err)

		return nil
	}
	defer resp.Body.Close()

	var result BuildSwapTransactionResponseBody

	json.NewDecoder(resp.Body).Decode(&result)

	return &result

}

func New(c *config.JupiterConfig) *Client {
	return &Client{config: c}
}
