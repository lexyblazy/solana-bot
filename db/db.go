package db

import (
	"database/sql"
	"log"
	"solana-bot/dexscreener"
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
		log.Println("MarkLogEventAsProcessed:", err)

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

func (s *SqlClient) GetTokensForProcessing(limit uint) []Token {
	var tokens []Token

	query := `select t."contractAddress" from  tokens t where t."lastProcessedAt" is null LIMIT ?`

	rows, err := s.db.Query(query, limit)

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
		"lastProcessedAt" = ?,
		symbol = ?,
		"marketCap" = ?,
		"pairCreatedAt" = ?
	 where "contractAddress" = ?`

	_, err := s.db.Exec(query, time.Now(), token.BaseToken.Symbol, token.MarketCap, time.UnixMilli(token.PairCreatedAt), token.BaseToken.Address)

	if err != nil {
		log.Printf("updateTokenMetadata: Failed to update token %s \n: err: %s", token.BaseToken.Address, err)
	}

}

func (s *SqlClient) UpdateTokenMetaData(tokens []dexscreener.TokensByAddress) {

	for _, token := range tokens {
		s.updateTokenMetadata(token)
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
