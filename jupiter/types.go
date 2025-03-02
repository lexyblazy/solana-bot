package jupiter

import (
	"fmt"
	"strconv"
)

type GetQuoteParams struct {
	InputMint   string
	OutputMint  string
	Amount      float64
	SlippageBps int
}

func (p GetQuoteParams) toQueryString() string {

	return fmt.Sprintf("inputMint=%s&outputMint=%s&amount=%s&slippageBps=%s",
		p.InputMint, p.OutputMint,
		strconv.Itoa(int(p.Amount)), strconv.Itoa(p.SlippageBps),
	)

}

type GetQuoteResponse struct {
	InputMint            string  `json:"inputMint"`
	InAmount             string  `json:"inAmount"`
	OutputMint           string  `json:"outputMint"`
	OutAmount            string  `json:"outAmount"`
	OtherAmountThreshold string  `json:"otherAmountThreshold"`
	SwapMode             string  `json:"swapMode"`
	SlippageBps          int     `json:"slippageBps"`
	PlatformFee          *string `json:"platformFee"`
	PriceImpactPct       string  `json:"priceImpactPct"`
	RoutePlan            []struct {
		SwapInfo struct {
			AmmKey     string `json:"ammKey"`
			Label      string `json:"label"`
			InputMint  string `json:"inputMint"`
			OutputMint string `json:"outputMint"`
			InAmount   string `json:"inAmount"`
			OutAmount  string `json:"outAmount"`
			FeeAmount  string `json:"feeAmount"`
			FeeMint    string `json:"feeMint"`
		} `json:"swapInfo"`
		Percent int `json:"percent"`
	} `json:"routePlan"`
	ScoreReport      *string `json:"scoreReport"`
	ContextSlot      int     `json:"contextSlot"`
	TimeTaken        float64 `json:"timeTaken"`
	SwapUsdValue     string  `json:"swapUsdValue"`
	SimplerRouteUsed bool    `json:"simplerRouteUsed"`
}

type BuildSwapTransactionRequestBody struct {
	QuoteResponse           GetQuoteResponse `json:"quoteResponse"`
	UserPublicKey           string           `json:"userPublicKey"`
	DynamicComputeUnitLimit bool             `json:"dynamicComputeUnitLimit"`
	DynamicSlippage         bool             `json:"dynamicSlippage"`
}

type BuildSwapTransactionResponseBody struct {
	SwapTransaction           string `json:"swapTransaction"`
	LastValidBlockHeight      int    `json:"lastValidBlockHeight"`
	PrioritizationFeeLamports int    `json:"prioritizationFeeLamports"`
	ComputeUnitLimit          int    `json:"computeUnitLimit"`
	PrioritizationType        struct {
		ComputeBudget struct {
			MicroLamports          int `json:"microLamports"`
			EstimatedMicroLamports int `json:"estimatedMicroLamports"`
		} `json:"computeBudget"`
	} `json:"prioritizationType"`
	SimulationSlot        int `json:"simulationSlot"`
	DynamicSlippageReport struct {
		SlippageBps                  int         `json:"slippageBps"`
		OtherAmount                  int         `json:"otherAmount"`
		SimulatedIncurredSlippageBps int         `json:"simulatedIncurredSlippageBps"`
		AmplificationRatio           interface{} `json:"amplificationRatio"`
		CategoryName                 string      `json:"categoryName"`
		HeuristicMaxSlippageBps      int         `json:"heuristicMaxSlippageBps"`
		RtseSlippageBps              int         `json:"rtseSlippageBps"`
		FailedTxnEstSlippage         int         `json:"failedTxnEstSlippage"`
		PriceMovementEstSlippage     int         `json:"priceMovementEstSlippage"`
		EmaEstSlippage               int         `json:"emaEstSlippage"`
	} `json:"dynamicSlippageReport"`
	SimulationError               interface{} `json:"simulationError"`
	AddressesByLookupTableAddress interface{} `json:"addressesByLookupTableAddress"`
}
