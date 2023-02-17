package status

// Status represents enum-like values received in the dispatcher from the RPC node.
// It includes a subset of well-known and likely failures the dispatcher may encounter.
type Status string

const (
	FailGasPrice        = "FAIL_LOW_GAS_PRICE"
	FailInsufficientGas = "FAIL_NO_GAS"
	FailNonce           = "FAIL_LOW_NONCE"
	Successful          = "SUCCESSFUL"
	Unknown             = "UNKNOWN"
)
