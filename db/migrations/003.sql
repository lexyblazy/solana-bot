-- UP
CREATE TABLE market_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, 
    timestamp DATETIME NOT NULL, 
    marketCap REAL NOT NULL DEFAULT 0,
    fdv REAL NOT NULL DEFAULT 0,
    liquidityUsd REAL NOT NULL DEFAULT 0,
    priceNative REAL NOT NULL DEFAULT 0,
    priceUsd REAL NOT NULL DEFAULT 0,
    contractAddress VARCHAR(255) NOT NULL
);

CREATE INDEX market_data_timestamp ON market_data("timestamp");
CREATE INDEX market_data_marketCap ON market_data("marketCap");
CREATE INDEX market_data_contractAddress ON market_data("contractAddress");

-- DOWN
DROP TABLE market_data