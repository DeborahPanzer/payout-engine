package sepa

import (
	"math/big"
	"strings"
)

func ValidateIBAN(rawIban string) bool {
	cleaned := strings.ReplaceAll(strings.ToUpper(strings.TrimSpace(rawIban)), " ", "")
	if len(cleaned) < 15 || len(cleaned) > 34 {
		return false
	}

	rearranged := cleaned[4:] + cleaned[:4]
	var numericBuilder strings.Builder

	for i := 0; i < len(rearranged); i++ {
		ch := rearranged[i]
		if ch >= '0' && ch <= '9' {
			numericBuilder.WriteByte(ch)
		} else if ch >= 'A' && ch <= 'Z' {
			val := int(ch-'A') + 10
			numericBuilder.WriteString(string(rune('0'+val/10)) + string(rune('0'+val%10)))
		} else {
			return false
		}
	}

	ibanInt, ok := new(big.Int).SetString(numericBuilder.String(), 10)
	if !ok {
		return false
	}
	mod := new(big.Int).Mod(ibanInt, big.NewInt(97))
	return mod.Int64() == 1
}
