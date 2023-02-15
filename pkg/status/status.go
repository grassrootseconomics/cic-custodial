package status

type Status string

const (
	FailGasPrice        = "FAIL_LOW_GAS_PRICE"
	FailInsufficientGas = "FAIL_NO_GAS"
	FailNonce           = "FAIL_LOW_NONCE"
	Successful          = "SUCCESSFUL"
	Unknown             = "UNKNOWN"
)
