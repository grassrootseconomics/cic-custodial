-- Otx tx type enum table
CREATE TABLE IF NOT EXISTS otx_tx_type (
  value TEXT PRIMARY KEY
);
INSERT INTO otx_tx_type (value) VALUES
('GIFT_GAS'),
('ACCOUNT_REGISTER'),
('GIFT_VOUCHER'),
('REFILL_GAS'),
('TRANSFER_VOUCHER');

-- Origin tx table
CREATE TABLE IF NOT EXISTS otx_sign (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tracking_id uuid NOT NULL,
    "type" TEXT REFERENCES otx_tx_type(value) NOT NULL,
    raw_tx TEXT NOT NULL,
    tx_hash TEXT NOT NULL,
    "from" TEXT NOT NULL,
    "data" TEXT NOT NULL,
    gas_price bigint NOT NULL,
    gas_limit bigint NOT NULL,
    transfer_value bigint NOT NULL,
    nonce int NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS tracking_id_idx ON otx_sign (tracking_id);
CREATE INDEX IF NOT EXISTS tx_hash_idx ON otx_sign (tx_hash);
CREATE INDEX IF NOT EXISTS from_idx ON otx_sign ("from");

-- Otx dispatch status enum table
-- Enforces referential integrity on the dispatch table
CREATE TABLE IF NOT EXISTS otx_dispatch_status_type (
  value TEXT PRIMARY KEY
);
INSERT INTO otx_dispatch_status_type (value) VALUES
('IN_NETWORK'),
('OBSOLETE'),
('SUCCESS'),
('FAIL_NO_GAS'),
('FAIL_LOW_NONCE'),
('FAIL_LOW_GAS_PRICE'),
('FAIL_UNKNOWN_RPC_ERROR'),
('REVERTED');

-- Dispatch status table
CREATE TABLE IF NOT EXISTS otx_dispatch (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    otx_id INT REFERENCES otx_sign(id),
    "status" TEXT REFERENCES otx_dispatch_status_type(value) NOT NULL,
    "block" bigint,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS status_idx ON otx_dispatch("status");
