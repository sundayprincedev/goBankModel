package currency

import "math"

type Currency struct {
	Name          string
	ScalingFactor float64
}

var Currencies = map[string]Currency{
	"USD": {Name: "US Dollar", ScalingFactor: 100},
	"NGN": {Name: "Nigerian Naira", ScalingFactor: 100},
	"EUR": {Name: "Euro", ScalingFactor: 100},
	"GBP": {Name: "British Pound", ScalingFactor: 100},
	"JPY": {Name: "Japanese Yen", ScalingFactor: 1},
	"CAD": {Name: "Canadian Dollar", ScalingFactor: 100},
	"AUD": {Name: "Australian Dollar", ScalingFactor: 100},
}

func CurrencyExists(currencyCode string) (Currency, bool) {
	val, ok := Currencies[currencyCode]
	return val, ok
}

func NormalizeAmountByRound(amount float64, curr Currency) int64 {
	scalingfactor := curr.ScalingFactor
	amountInt := int64(math.Round(amount * scalingfactor))
	return amountInt
}

func NormalizeAmountByFloor(amount float64, curr Currency) int64 {
	scalingfactor := curr.ScalingFactor
	amountInt := int64(math.Floor(amount * scalingfactor))
	return amountInt
}

func DenormalizeAmount(amount int64, curr Currency) float64 {
	amountToFloat := float64(amount)
	amountFloat := float64(amountToFloat / curr.ScalingFactor)
	return amountFloat
}
