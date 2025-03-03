-- UP
CREATE TABLE swap_orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, 
    createdAt DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    lastProcessedAt DATETIME DEFAULT NULL,
    executedAt DATETIME DEFAULT NULL,
    fromToken VARCHAR(255) NOT NULL,
    toToken VARCHAR(255) NOT NULL,
    txHash VARCHAR(255) DEFAULT NULL,
    rules TEXT DEFAULT  NULL,
    amountDetails TEXT DEFAULT NULL
);

CREATE INDEX swap_orders_fromToken ON swap_orders("fromToken");
CREATE INDEX swap_orders_toToken ON swap_orders("toToken");
CREATE INDEX swap_orders_txHash ON swap_orders("txHash");

-- DOWN
DROP TABLE swap_orders