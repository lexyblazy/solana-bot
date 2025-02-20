package helius

type GetParsedTxReqBody struct {
	Transactions []string `json:"transactions"`
}

type ParsedTx struct {
	Signature string `json:"signature"`
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
	JsonRPC string        `json:"jsonrpc"`
	Method  string        `json:"logsNotification"`
	Params  messageParams `json:"params"`
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
