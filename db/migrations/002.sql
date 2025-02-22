-- UP
CREATE TABLE tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, 
    contractAddress VARCHAR(255) NOT NULL,
    createdAt DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    lastProcessedAt DATETIME DEFAULT NULL
);

CREATE UNIQUE INDEX tokens_unique_contractAddress  ON tokens("contractAddress");
CREATE INDEX tokens_marketCap  ON tokens("marketCap");

ALTER TABLE tokens ADD symbol VARCHAR
ALTER TABLE tokens ADD marketCap REAL
ALTER TABLE tokens ADD pairCreatedAt DATETIME



-- DOWN

DROP TABLE tokens