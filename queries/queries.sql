--name: write-key-pair
-- Save hex encoded private key
-- $1: public_key
-- $2: private_key
INSERT INTO keystore(public_key, private_key) VALUES($1, $2) RETURNING id

--name: load-key-pair
-- Load saved key pair
-- $1: public_key
SELECT private_key FROM keystore WHERE public_key=$1

--name: create-otx
-- Create a new locally originating tx
-- $1: tracking_id
-- $2: type
-- $3: raw_tx
-- $4: tx_hash
-- $5: from
-- $6: data
-- $7: gas_price
-- $8: gas_limit
-- $9: transfer_value
-- $10: nonce
INSERT INTO otx_sign(
    tracking_id,
    "type",
    raw_tx,
    tx_hash,
    "from",
    "data",
    gas_price,
    gas_limit,
    transfer_value,
    nonce
) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id

--name: create-dispatch-status
-- Create a new dispatch status
-- $1: otx_id
-- $2: status
INSERT INTO otx_dispatch(
    otx_id,
    "status"
) VALUES($1, $2) RETURNING id

--name: update-chain-status
-- Updates the status of the dispatched tx with the chain mine status
-- $1: tx_hash
-- $2: status
-- $3: block
UPDATE otx_dispatch SET "status" = $2, "block" = $3 WHERE otx_dispatch.id = (
    SELECT otx_dispatch.id FROM otx_dispatch
    INNER JOIN otx_sign ON otx_dispatch.otx_id = otx_sign.id
    WHERE otx_sign.tx_hash = $1
    AND otx_dispatch.status = 'IN_NETWORK'
)

--name: get-tx-status-by-tracking-id
-- Gets tx status's from possible multiple txs with the same tracking_id
-- $1: tracking_id
SELECT otx_sign.type, otx_sign.tx_hash, otx_sign.transfer_value, otx_sign.created_at, otx_dispatch.status FROM otx_sign
INNER JOIN otx_dispatch ON otx_sign.id = otx_dispatch.otx_id
WHERE otx_sign.tracking_id = $1

-- TODO: Scroll by status type with cursor pagination