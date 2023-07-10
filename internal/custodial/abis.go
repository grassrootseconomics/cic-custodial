package custodial

import "github.com/grassrootseconomics/w3-celo-patch"

const (
	Approve      = "approve"
	Check        = "check"
	GiveTo       = "giveTo"
	MintTo       = "mintTo"
	NextTime     = "nextTime"
	Register     = "register"
	Transfer     = "transfer"
	TransferFrom = "transferFrom"
)

// Define common smart contrcat ABI's that can be injected into the system container.
// Any relevant function signature that will be used by the custodial system can be defined here.
func initAbis() map[string]*w3.Func {
	return map[string]*w3.Func{
		Approve:  w3.MustNewFunc("approve(address, uint256)", "bool"),
		Check:    w3.MustNewFunc("check(address)", "bool"),
		GiveTo:   w3.MustNewFunc("giveTo(address)", "uint256"),
		MintTo:   w3.MustNewFunc("mintTo(address, uint256)", "bool"),
		NextTime: w3.MustNewFunc("nextTime(address)", "uint256"),
		Register: w3.MustNewFunc("register(address)", ""),
		Transfer: w3.MustNewFunc("transfer(address,uint256)", "bool"),
	}
}
