-- Keystore queries

--name: write-key-pair
-- Save hex encoded private key
-- $1: public_key
-- $2: private_key
INSERT INTO keystore(public_key, private_key) VALUES($1, $2) RETURNING id

--name: load-key-pair
-- Load saved key pair
-- $1: public_key
SELECT private_key FROM keystore WHERE id=$1

-- OTX queries

--name: create-otx
-- Create a new locally originating tx
-- $1: raw_tx
-- $2: tx_hash
-- $3: from
-- $4: data
-- $5: gas_price
-- $6: nonce
-- $7: tracking_id
INSERT INTO otx(
    raw_tx,
    tx_hash,
    from,
    data,
    gas_price,
    nonce,
    tracking_id
) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id


-- Dispatch status queries

--name: create-dispatch-status
-- Create a new dispatch status
-- $1: otx_id
-- $2: status
-- £3: tracking_id
INSERT INTO otx(
    otx_id,
    status,
    tracking_id
) VALUES($1, $2, $3) RETURNING id