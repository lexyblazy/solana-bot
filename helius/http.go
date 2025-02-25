package helius

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"solana-bot/config"
)

type HttpClient struct {
	config *config.HeliusConfig
}

func (h *HttpClient) GetParsedTxs(signatures []string) ([]ParsedTx, error) {

	url := fmt.Sprintf("%s/transactions?api-key=%s", h.config.BaseApiUrl, h.config.ApiKey)

	body, err := json.Marshal(GetParsedTxReqBody{
		Transactions: signatures,
	})

	if err != nil {

		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))

	if err != nil {
		return nil, fmt.Errorf("failed to create new request %s", err)
	}

	req.Header.Set("Authorization", "Basic username:password")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("GetParsedTxs: Request failed: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println()

		return nil, fmt.Errorf("GetParsedTxs: Failed to retrieve tx data %d", resp.StatusCode)
	}

	var parsedTxs []ParsedTx

	json.NewDecoder(resp.Body).Decode(&parsedTxs)

	return parsedTxs, nil

}

// returns the balance in lamports
func (h *HttpClient) GetBalance(address string) int {
	url := fmt.Sprintf("%s?api-key=%s", h.config.RpcUrl, h.config.ApiKey)

	reqBody, err := json.Marshal(GetBalanceRequestBody{
		BaseRPCBody: BaseRPCBody{
			ID:      "1",
			JsonRPC: "2.0",
			Method:  "getBalance",
		},
		Params: []string{address},
	})

	if err != nil {
		log.Println("GetBalance: Failed to marshal body", err)

		return 0
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))

	if err != nil {
		log.Println("GetBalance: Failed to Get url", err)

		return 0
	}

	defer resp.Body.Close()

	var responseBody GetBalanceResponseBody

	json.NewDecoder(resp.Body).Decode(&responseBody)

	return responseBody.Result.Value

}

func NewHttpClient(c *config.HeliusConfig) *HttpClient {
	return &HttpClient{config: c}
}
