package util

import (
	"fmt"

	"github.com/IfanTsai/go-lib/randutils"
)

// RandomOwner generates a random owner name.
func RandomOwner() string {
	return randutils.RandomString(6)
}

// RandomMoney generates a random amount of money.
func RandomMoney() int64 {
	return randutils.RandomInt(0, 1000)
}

// RandomCurrency generates a random currency code.
func RandomCurrency() string {
	currencies := []string{EUR, USD, CAD}

	return currencies[randutils.RandomInt(0, int64(len(currencies)-1))]
}

// RandomEmail generates a random email.
func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", randutils.RandomString(6))
}
