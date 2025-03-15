package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"solana-bot/dexscreener"
	"solana-bot/utils"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SqlClient struct {
	db *sql.DB
}

func (s *SqlClient) Close() {
	s.db.Close()
}

func (s *SqlClient) InsertLog(signature string) {

	query := `insert into rpc_logs(signature) values(?)`

	_, err := s.db.Exec(query, signature)

	if err != nil {
		log.Println("InsertLog:", err)

		return
	}

}

func (s *SqlClient) InsertNewToken(contractAddress string, signature string) {

	_, err := s.db.Exec(`insert into tokens("contractAddress") values(?)`, contractAddress)

	if err != nil {
		log.Println("InsertNewTokens:", err)

		return
	}

	// mark event that corresponds to this signature as processed
	s.UpdateLogEventAsProcessed(signature)

}

func (s *SqlClient) UpdateLogEventAsProcessed(signature string) {
	_, err := s.db.Exec(`update rpc_logs set "processedAt" = ? where signature = ? and "processedAt" is null`, time.Now(), signature)

	if err != nil {
		log.Println("UpdateLogEventAsProcessed:", err)

		return
	}

}

func toInterfaceSlice(values []string) []interface{} {
	var result []interface{}
	for _, name := range values {
		result = append(result, name)
	}
	return result
}

func makePlaceHolders(length int) []string {
	placeholders := make([]string, length)

	for i := 0; i < length; i++ {
		placeholders[i] = "?"
	}

	return placeholders
}

func (s *SqlClient) UpdateTokensAsProcessed(addresses []string) {

	placeholders := makePlaceHolders(len(addresses))

	query := fmt.Sprintf("update tokens set lastProcessedAt = ? where contractAddress in (%s)", strings.Join(placeholders, ","))

	var params []interface{}
	params = append(params, time.Now().UnixMilli()) // add the lastProcessedAt to the values list
	params = append(params, toInterfaceSlice(addresses)...)
	_, err := s.db.Exec(query, params...)

	if err != nil {
		log.Println("UpdateTokensAsProcessed Query:", err)

		return
	}

}

func (s *SqlClient) DeleteLogs() {

	query := `delete from rpc_logs  where "processedAt" is not null`

	result, err := s.db.Exec(query)

	if err != nil {
		log.Println("DeleteProcessedLogs:", err)

		return
	}

	count, err := result.RowsAffected()

	if err != nil {
		log.Println("DeleteProcessedLogs:", err)

		return
	}

	if count > 0 {
		log.Printf("Deleted %d processed logs \n", count)
	}

}

func (s *SqlClient) GetUnProcessedLogs(limit uint) []RpcLogEntity {

	var rpcLogs []RpcLogEntity

	query := `select id, signature, createdAt, processedAt from  rpc_logs r where r."processedAt" is null LIMIT ?`

	rows, err := s.db.Query(query, limit)

	if err != nil {
		log.Println("GetUnProcessedLogs:", err)

		return rpcLogs
	}

	for rows.Next() {

		var rpcLog RpcLogEntity
		err = rows.Scan(&rpcLog.Id, &rpcLog.Signature, &rpcLog.CreatedAt, &rpcLog.ProcessedAt)

		if err != nil {
			log.Println("GetUnProcessedLogs:", err)
			break
		}
		rpcLogs = append(rpcLogs, rpcLog)
	}

	return rpcLogs

}

func (s *SqlClient) GetScamTokens(minMarketCap int, minAgeHours int) []string {

	threshold := time.Now().UnixMilli() - int64(minAgeHours)*time.Hour.Milliseconds()

	query := `select t.contractAddress from tokens t where t.marketCap < ? AND t.pairCreatedAt < ? AND t.lastProcessedAt is not null`

	rows, err := s.db.Query(query, minMarketCap, threshold)

	if err != nil {
		log.Println("GetScamTokens:", err)
	}

	var addresses []string
	for rows.Next() {
		var contractAddress string
		rows.Scan(&contractAddress)
		addresses = append(addresses, contractAddress)
	}

	return addresses
}

func (s *SqlClient) DeleteTokens(addresses []string) {
	placeholders := makePlaceHolders(len(addresses))

	// delete the tokens and their market data
	tx, err := s.db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})

	if err != nil {
		log.Println("DeleteTokens: Failed to begin tx", err)
		return
	}

	// delete market data query
	query1 := fmt.Sprintf(`delete from market_data where contractAddress IN (%s)`, strings.Join(placeholders, ","))
	result1, err := tx.Exec(query1, toInterfaceSlice(addresses)...)

	if err != nil {
		log.Print("DeleteTokens: Failed to delete market data ")
		return
	}

	c1, _ := result1.RowsAffected()
	log.Printf("DeleteTokens: Deleted %d market_data records \n", c1)

	// delete token query
	query2 := fmt.Sprintf(`delete from tokens where contractAddress IN (%s)`, strings.Join(placeholders, ","))
	result2, err := tx.Exec(query2, toInterfaceSlice(addresses)...)

	if err != nil {
		log.Print("DeleteTokens: Failed to delete tokens ")

		tx.Rollback()
		return
	}

	c2, _ := result2.RowsAffected()
	log.Printf("DeleteTokens: Deleted %d token records \n", c2)

	err = tx.Commit()

	if err != nil {
		log.Println("DeleteTokens: Failed to commit tx", err)
		return
	}

	log.Printf("DeleteTokens: Deleted %d tokens \n", len(addresses))
}

func (s *SqlClient) GetTopTokens(marketCap int) []TokenEntity {
	var tokens []TokenEntity

	query := `select t."contractAddress" from  tokens t where t."marketCap" > ?`

	rows, err := s.db.Query(query, marketCap)

	if err != nil {
		log.Println("GetTopTokens:", err)

		return tokens
	}

	for rows.Next() {

		var token TokenEntity
		err = rows.Scan(&token.ContractAddress)

		if err != nil {
			log.Println("GetTokensForProcessing:", err)
			break
		}

		tokens = append(tokens, token)
	}

	return tokens

}

func (s *SqlClient) GetTokensForProcessing(limit int, frequencyMinutes int, minMarketCap int) []TokenEntity {
	var tokens []TokenEntity

	minThreshold := time.Now().UnixMilli() - int64(frequencyMinutes)*time.Minute.Milliseconds()

	query := `select t."contractAddress" from  tokens t where t."marketCap" < ? AND (t."lastProcessedAt" is null OR t.lastProcessedAt < ?) LIMIT ?`

	rows, err := s.db.Query(query, minMarketCap, minThreshold, limit)

	if err != nil {
		log.Println("GetTokensForProcessing:", err)

		return tokens
	}

	for rows.Next() {

		var token TokenEntity
		err = rows.Scan(&token.ContractAddress)

		if err != nil {
			log.Println("GetTokensForProcessing:", err)
			break
		}

		tokens = append(tokens, token)
	}

	return tokens
}

func (s *SqlClient) updateTokenMetadata(token dexscreener.TokensByAddress) {
	query := `
	update tokens set
		symbol = ?,
		marketCap = ?,
		"pairCreatedAt" = ?
	 where "contractAddress" = ?`

	_, err := s.db.Exec(query, token.BaseToken.Symbol, token.MarketCap, token.PairCreatedAt, token.BaseToken.Address)

	if err != nil {
		log.Printf("updateTokenMetadata: Failed to update token %s \n: err: %s", token.BaseToken.Address, err)
	}
}

func (s *SqlClient) insertMarketData(token dexscreener.TokensByAddress) {
	query := `insert into market_data("timestamp","marketCap", "fdv", "liquidityUsd", "priceNative", "priceUsd", "contractAddress")
	 values(?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, time.Now().UnixMilli(), token.MarketCap, token.Fdv, token.Liquidity.Usd, token.PriceNative, token.PriceUsd, token.BaseToken.Address)

	if err != nil {
		log.Printf("insertMarketData: Failed to insert marketData %s \n: err: %s", token.BaseToken.Address, err)
	}

}

func (s *SqlClient) GetTokensByContractAddress(addresses []string) []TokenEntity {
	placeholders := makePlaceHolders(len(addresses))
	var tokens []TokenEntity

	query := fmt.Sprintf(`select t.contractAddress, t.symbol, t.pairCreatedAt from tokens t where t.contractAddress IN (%s)`, strings.Join(placeholders, ","))

	rows, err := s.db.Query(query, toInterfaceSlice(addresses)...)

	if err != nil {
		log.Println("GetTokensByContractAddress:", err)

		return tokens
	}

	for rows.Next() {
		var token TokenEntity
		rows.Scan(&token.ContractAddress, &token.Symbol, &token.PairCreatedAt)

		tokens = append(tokens, token)
	}

	return tokens
}

func (s *SqlClient) GetTokenMarketData(address string) []MarketDataEntity {

	var marketData []MarketDataEntity

	rows, err := s.db.Query(`select timestamp, marketCap, liquidityUsd from market_data md where md.contractAddress = ? order by md.marketCap desc`, address)

	if err != nil {

		log.Println("GetTokenMarketData:", err)
		return marketData
	}

	for rows.Next() {
		var m MarketDataEntity

		rows.Scan(&m.Timestamp, &m.MarketCap, &m.LiquidityUsd)

		marketData = append(marketData, m)
	}

	return marketData

}

func (s *SqlClient) UpdateTokenData(tokens []dexscreener.TokensByAddress) {

	for _, token := range tokens {
		go s.insertMarketData(token)
		go s.updateTokenMetadata(token)

	}
}

func (s *SqlClient) InsertBuyOrder(st *SwapTradeEntity) {

	query := `insert into swap_orders("fromToken", "toToken", "amountDetails") VALUES (?, ?, ?)`

	_, err := s.db.Exec(query, st.FromToken, st.ToToken, utils.ToString(st.AmountDetails))

	if err != nil {
		log.Println("InsertBuyOrder:", err)

		return
	}

	log.Print("InsertBuyOrder DONE!")

}

func (s *SqlClient) InsertSellOrder(st *SwapTradeEntity) {
	query := `insert into swap_orders("fromToken", "toToken", "rules") VALUES (?, ?, ?)`

	_, err := s.db.Exec(query, st.FromToken, st.ToToken, utils.ToString(st.Rules))

	if err != nil {
		log.Println("InsertSellOrder:", err)

		return
	}

	log.Print("InsertSellOrder DONE!")

}

func (s *SqlClient) GetPendingTrades() []SwapTradeEntity {
	query := `select id, fromToken, toToken, amountDetails, rules from swap_orders sp where sp."executedAt" is null`

	rows, err := s.db.Query(query)

	if err != nil {
		log.Print("GetPendingTrades: dbQuery Error", err)
	}

	var trades []SwapTradeEntity

	for rows.Next() {
		var trade SwapTradeEntity
		rows.Scan(&trade.Id, &trade.FromToken, &trade.ToToken, &trade.AmountDetails, &trade.Rules)

		trades = append(trades, trade)
	}

	return trades

}

func New(dbPath string) *SqlClient {
	db, err := sql.Open("sqlite3", dbPath)

	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatal("unable to use data source name", err)
	}

	return &SqlClient{db: db}

}
