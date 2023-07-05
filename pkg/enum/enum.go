package enum

type (
	// OtxStatus represents enum-like values received in the dispatcher from the RPC node or Network callback.
	// It includes a subset of well-known and likely failures the dispatcher may encounter.
	OtxStatus string
	// OtxType reprsents the specific type of signed transaction.
	OtxType string
)

// NOTE: These values must also be inserted/updated into db to enforce referential integrity.
const (
	IN_NETWORK             OtxStatus = "IN_NETWORK"
	OBSOLETE               OtxStatus = "OBSOLETE"
	SUCCESS                OtxStatus = "SUCCESS"
	FAIL_NO_GAS            OtxStatus = "FAIL_NO_GAS"
	FAIL_LOW_NONCE         OtxStatus = "FAIL_LOW_NONCE"
	FAIL_LOW_GAS_PRICE     OtxStatus = "FAIL_LOW_GAS_PRICE"
	FAIL_UNKNOWN_RPC_ERROR OtxStatus = "FAIL_UNKNOWN_RPC_ERROR"
	REVERTED               OtxStatus = "REVERTED"

	ACCOUNT_REGISTER OtxType = "ACCOUNT_REGISTER"
	REFILL_GAS       OtxType = "REFILL_GAS"
	TRANSFER_AUTH    OtxType = "TRANSFER_AUTHORIZATION"
	TRANSFER_VOUCHER OtxType = "TRANSFER_VOUCHER"
)
