package crypto

import (
	"fmt"
	"math/big"
)

func FromWei(value string, decimals int) (*big.Rat, error) {
	i := new(big.Int)
	if _, ok := i.SetString(value, 10); !ok {
		return nil, fmt.Errorf("invalid integer: %s", value)
	}

	denom := new(big.Int).Exp(
		big.NewInt(10),
		big.NewInt(int64(decimals)),
		nil,
	)

	return new(big.Rat).SetFrac(i, denom), nil
}

func ParseCryptoAmount(value string) string {
	r, err := FromWei(value, 18)
	if err != nil {
		return "0"
	}
	return r.FloatString(6)
}
