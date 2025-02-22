package db

import (
	"database/sql"
	"fmt"
	"log"
	"solana-bot/dexscreener"
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

func (s *SqlClient) UpdateTokensAsProcessed(addresses []string) {

	placeholders := make([]string, len(addresses))

	for i := range addresses {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf("update tokens set lastProcessedAt = ? where contractAddress in (%s)", strings.Join(placeholders, ","))

	var params []interface{}
	params = append(params, time.Now().UnixMilli()) // add the lastProcessedAt to the values list
	params = append(params, toInterfaceSlice(addresses)...)
	result, err := s.db.Exec(query, params...)

	if err != nil {
		log.Println("UpdateTokensAsProcessed Query:", err)

		return
	}

	count, err := result.RowsAffected()

	if err != nil {
		log.Println("UpdateTokensAsProcessed RowsAffected:", err)

		return
	}

	log.Printf("UpdateTokensAsProcessed: updated %d tokens \n", count)

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

	log.Printf("Deleted %d processed logs \n", count)

}

func (s *SqlClient) GetUnProcessedLogs(limit uint) []RpcLog {

	var rpcLogs []RpcLog

	query := `select id, signature, createdAt, processedAt from  rpc_logs r where r."processedAt" is null LIMIT ?`

	rows, err := s.db.Query(query, limit)

	if err != nil {
		log.Println("GetUnProcessedLogs:", err)

		return rpcLogs
	}

	for rows.Next() {

		var rpcLog RpcLog
		err = rows.Scan(&rpcLog.Id, &rpcLog.Signature, &rpcLog.CreatedAt, &rpcLog.ProcessedAt)

		if err != nil {
			log.Println("GetUnProcessedLogs:", err)
			break
		}
		rpcLogs = append(rpcLogs, rpcLog)
	}

	return rpcLogs

}

func (s *SqlClient) GetTokensForProcessing(limit int, frequencyMinutes int) []Token {
	var tokens []Token

	minThreshold := time.Now().UnixMilli() - int64(frequencyMinutes)*time.Minute.Milliseconds()

	query := `select t."contractAddress" from  tokens t where t."lastProcessedAt" is null or t.lastProcessedAt < ? LIMIT ?`

	rows, err := s.db.Query(query, minThreshold, limit)

	if err != nil {
		log.Println("GetTokensForProcessing:", err)

		return tokens
	}

	for rows.Next() {

		var token Token
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

func (s *SqlClient) UpdateTokenData(tokens []dexscreener.TokensByAddress) {

	for _, token := range tokens {
		go s.insertMarketData(token)
		go s.updateTokenMetadata(token)

	}
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
