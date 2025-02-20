-- UP
CREATE TABLE rpc_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, 
    signature VARCHAR(255) NOT NULL,
    createdAt DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    processedAt DATETIME DEFAULT NULL
);

CREATE UNIQUE INDEX rpc_logs_unique_signature ON rpc_logs("signature");
CREATE INDEX rpc_logs_createdAt ON rpc_logs("createdAt");
CREATE INDEX rpc_logs_processedAt ON rpc_logs("processedAt");


-- DOWN
DROP TABLE rpc_logs