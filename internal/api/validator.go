package api

import (
	"github.com/celo-org/celo-blockchain/common"
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	ValidatorProvider *validator.Validate
}

func (v *Validator) Validate(i interface{}) error {
	if err := v.ValidatorProvider.Struct(i); err != nil {
		return err
	}
	return nil
}

func EthChecksumValidator(fl validator.FieldLevel) bool {
	addr, err := common.NewMixedcaseAddressFromString(fl.Field().String())
	if err != nil {
		return false
	}

	return addr.ValidChecksum()
}
