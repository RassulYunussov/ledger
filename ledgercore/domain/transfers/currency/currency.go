package currency

// https://en.wikipedia.org/wiki/List_of_circulating_currencies
// functionality to convert incoming amount into fractional and back
// assuming default number to basic as 100
// all the other currencies should be declared explicitly

const defaultNumberToBasic = 100

var currencyNumberToBasic = map[string]int32{
	"BHD": 1000,
	"BND": 10,
	"SGD": 10,
	"CNY": 10,
	"IRR": 1,
	"IQD": 1000,
	"KWD": 1000,
	"LYD": 1000,
	"MGA": 5,
	"MRU": 5,
	"OMR": 1000,
	"TND": 1000,
	"VES": 1,
	"VND": 10,
}

func toFractional(amount int64, fractional int32, numberToBasic int32) int64 {
	return amount*int64(numberToBasic) + int64(fractional)
}

func fromFractional(amount int64, numberToBasic int32) (int64, int32) {
	return amount / int64(numberToBasic), int32(amount % int64(numberToBasic))
}

func getNumberToBasic(currency string) int32 {
	if v, ok := currencyNumberToBasic[currency]; ok {
		return v
	}
	return defaultNumberToBasic
}

func GetAmountInFractional(amount int64, fractional int32, currency string) int64 {
	numberToBasic := getNumberToBasic(currency)
	return toFractional(amount, fractional, numberToBasic)
}

func GetAmountFromFractional(amount int64, currency string) (int64, int32) {
	numberToBasic := getNumberToBasic(currency)
	return fromFractional(amount, numberToBasic)
}
