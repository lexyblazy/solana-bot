package helius

type GetParsedTxReqBody struct {
	Transactions []string `json:"transactions"`
}

type BaseRPCBody struct {
	ID      int `json:"id"`
	JsonRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
}

type ParsedTx struct {
	Signature    string `json:"signature"`
	Instructions []struct {
		Data              string   `json:"data"`
		ProgramId         string   `json:"programId"`
		Accounts          []string `json:"accounts"`
		InnerInstructions []struct {
			Data      string   `json:"data"`
			ProgramId string   `json:"programId"`
			Accounts  []string `json:"accounts"`
		} `json:"innerInstructions"`
	} `json:"instructions"`
}

type LogSubscribeMessage struct {
	BaseRPCBody
	Params messageParams `json:"params"`
}

type messageParams struct {
	Result       result `json:"result"`
	Subscription uint32 `json:"subscription"`
}

type result struct {
	Context struct {
		Slot int32 `json:"slot"`
	} `json:"context"`

	Value struct {
		Signature string   `json:"signature"`
		Logs      []string `json:"logs"`
	} `json:"value"`
}

type GetBalanceRequestBody struct {
	BaseRPCBody
	Params []string `json:"params"`
}

type GetBalanceResponseBody struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		Context struct {
			APIVersion string `json:"apiVersion"`
			Slot       int    `json:"slot"`
		} `json:"context"`
		Value int `json:"value"`
	} `json:"result"`
	ID string `json:"id"`
}
