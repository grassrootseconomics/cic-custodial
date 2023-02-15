-- Origin tx table
CREATE TABLE IF NOT EXISTS otx (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tracking_id TEXT NOT NULL,
    "type" TEXT NOT NULL,
    raw_tx TEXT NOT NULL,
    tx_hash TEXT NOT NULL,
    "from" TEXT NOT NULL,
    "data" TEXT NOT NULL,
    gas_price bigint NOT NULL,
    nonce int NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS tx_hash_idx ON otx USING hash(tx_hash);
CREATE INDEX IF NOT EXISTS from_idx ON otx USING hash("from");

-- Dispatch status table
CREATE TABLE IF NOT EXISTS dispatch (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    otx_id INT REFERENCES otx(id),
    "status" TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
