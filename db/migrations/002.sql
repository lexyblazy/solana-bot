-- UP
CREATE TABLE tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, 
    contractAddress VARCHAR(255) NOT NULL,
    createdAt DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    lastProcessedAt DATETIME DEFAULT NULL
);

CREATE UNIQUE INDEX tokens_unique_contractAddress  ON tokens("contractAddress");

ALTER TABLE tokens ADD symbol VARCHAR
ALTER TABLE tokens ADD marketCap REAL
ALTER TABLE tokens ADD pairCreatedAt DATETIME

CREATE INDEX tokens_marketCap  ON tokens("marketCap");
CREATE INDEX tokens_symbol ON tokens("symbol");


-- DOWN

DROP TABLE tokens