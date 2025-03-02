The general idea behind the Engine

### 1. Watch the Solana Blockchain for Raydium Liquidity Pool Migration events

Subscribe to the `logsSubscribe` event (via Helius websocket) to receive events, see `logsSubscribe.json`
Check the log messages, The message of concern in the logs is `"Program log: initialize2: InitializeInstruction2"`.
Once this message is found, we save the signature of that log to the `"rpc_logs"` table

```go
// subscribe to helius websocket streaming
	go e.hs.SubscribeToLogs()
	go e.hs.ReadMessages()

	for msg := range e.hs.GetMessageChannel() {
		go e.handleLogSubscribeMessage(msg)
	}

```



### 2. Run a job to process these log events.

Processing the logs means fetching the contractAddress that was involved in the event. By using the Helius transactions
API, we can retrieve the corresponding contractAddresses for the signatures. we save the contractAddresses in the `"tokens"` table.

```go
    go e.ProcessLogs()
```


### 3. Run a job to refresh the market data for the tokens

This job uses [Dexscreener APIs](https://docs.dexscreener.com/api/reference#tokens-v1-chainid-tokenaddresses) to fetch metadata (e.g symbol, creationDate) and marketData (e.g marketCap, fdv, liquidity etc)  for the tokens. 
Two things are done with the data returned from Dexscreener:
 - Update the `token` record in the database with the `symbol`, `creationDate` and `marketCap`
 - Insert a `market_data` record which is intended for tracking market data of the token

```go
    // targets tokens above a certain marketCap
    go e.RefreshTopTokensMetadata()
    // target tokens below the set marketCap
	go e.RefreshTokensMetadata()
```

### 4. Run jobs to delete Processed log events and scam tokens

```go
	go e.DeleteProcessedLogs()
    /* the rules for this job can be configured via the config.json file,
    current rule is remove tokens that are older than 48 hours AND a marketCap LESS THAN $50,000 */
	go e.RemoveScamTokens() 

```

