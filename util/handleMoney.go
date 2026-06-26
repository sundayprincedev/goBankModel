package util

import "fmt"

func HandleMoney(fromBal int64, toBal int64, amount int64) (int64, int64, error) {
	aNew := fromBal - amount
	if aNew < 0 {
		return fromBal, toBal, fmt.Errorf("error in money transfer")
	}

	return fromBal - amount, toBal + amount, nil
}
