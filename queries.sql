-- Keystore queries

--name: write-key-pair
-- Save hex encoded private key
-- $1: public_key
-- $2: private_key
INSERT INTO keystore(public_key, private_key) VALUES($1, $2) RETURNING id

--name: load-key-pair
-- Load saved key pair
-- $1: public_key
SELECT private_key FROM keystore WHERE public_key=$1

-- OTX queries

--name: create-otx
-- Create a new locally originating tx
-- $1: tracking_id
-- $2: type
-- $3: raw_tx
-- $4: tx_hash
-- $5: from
-- $6: data
-- $7: gas_price
-- $8: nonce
INSERT INTO otx(
    tracking_id,
    "type",
    raw_tx,
    tx_hash,
    "from",
    "data",
    gas_price,
    nonce
) VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id


-- Dispatch status queries

--name: create-dispatch-status
-- Create a new dispatch status
-- $1: otx_id
-- $2: status
INSERT INTO dispatch(
    otx_id,
    "status"
) VALUES($1, $2) RETURNING id