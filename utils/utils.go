package utils

import (
	"fmt"
	"math/big"
)

func HexToDec(hex string) (*big.Int, error) {
	value := new(big.Int)
	value.SetString(hex[2:], 16)
	return value, nil
}

func IntToHex[T int | int64](value T) string {
	return fmt.Sprintf("0x%x", value)
}

func DecimalStringToHexString(decimalValue string) (string, error) {
	value := new(big.Int)
	_, success := value.SetString(decimalValue, 10)
	if !success {
		return "", fmt.Errorf("invalid decimal value: %s", decimalValue)
	}
	hexValue := fmt.Sprintf("0x%x", value)
	return hexValue, nil
}
