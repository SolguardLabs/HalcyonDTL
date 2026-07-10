package main

import "fmt"

const (
	BpsScale     int64 = 10_000
	PPMScale     int64 = 1_000_000
	RayScale     int64 = 1_000_000_000
	DefaultAsset       = "hUSD"
)

type Amount int64

func NewAmount(value int64) (Amount, error) {
	if value < 0 {
		return 0, fmt.Errorf("amount cannot be negative: %d", value)
	}
	return Amount(value), nil
}

func MustAmount(value int64) Amount {
	amount, err := NewAmount(value)
	if err != nil {
		panic(err)
	}
	return amount
}

func (a Amount) Int64() int64 {
	return int64(a)
}

func (a Amount) IsZero() bool {
	return a == 0
}

func (a Amount) Positive() bool {
	return a > 0
}

func (a Amount) Add(other Amount) (Amount, error) {
	if int64(a) > maxInt64-int64(other) {
		return 0, fmt.Errorf("amount overflow")
	}
	return a + other, nil
}

func (a Amount) MustAdd(other Amount) Amount {
	out, err := a.Add(other)
	if err != nil {
		panic(err)
	}
	return out
}

func (a Amount) Sub(other Amount) (Amount, error) {
	if other > a {
		return 0, fmt.Errorf("amount underflow: %d - %d", a, other)
	}
	return a - other, nil
}

func (a Amount) MustSub(other Amount) Amount {
	out, err := a.Sub(other)
	if err != nil {
		panic(err)
	}
	return out
}

func (a Amount) Clamp(max Amount) Amount {
	if a > max {
		return max
	}
	return a
}

func (a Amount) Min(other Amount) Amount {
	if a < other {
		return a
	}
	return other
}

func (a Amount) Max(other Amount) Amount {
	if a > other {
		return a
	}
	return other
}

func (a Amount) String() string {
	return fmt.Sprintf("%d", a)
}

func (a Amount) MulBps(bps int64) (Amount, error) {
	return mulDivAmount(a, bps, BpsScale)
}

func (a Amount) MustMulBps(bps int64) Amount {
	out, err := a.MulBps(bps)
	if err != nil {
		panic(err)
	}
	return out
}

func (a Amount) MulPPM(ppm int64) (Amount, error) {
	return mulDivAmount(a, ppm, PPMScale)
}

func (a Amount) MustMulPPM(ppm int64) Amount {
	out, err := a.MulPPM(ppm)
	if err != nil {
		panic(err)
	}
	return out
}

func SignedAmountFromPPM(notional Amount, ppm int64) (int64, error) {
	if notional == 0 || ppm == 0 {
		return 0, nil
	}
	value := int64(notional)
	if absInt64(ppm) > maxInt64/value {
		return 0, fmt.Errorf("signed funding overflow")
	}
	return value * ppm / PPMScale, nil
}

func mulDivAmount(amount Amount, numerator int64, denominator int64) (Amount, error) {
	if denominator <= 0 {
		return 0, fmt.Errorf("invalid denominator")
	}
	if numerator < 0 {
		return 0, fmt.Errorf("negative numerator")
	}
	if amount == 0 || numerator == 0 {
		return 0, nil
	}
	if int64(amount) > maxInt64/numerator {
		return 0, fmt.Errorf("amount multiplication overflow")
	}
	return Amount(int64(amount) * numerator / denominator), nil
}

func absInt64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}

func clampInt64(value int64, min int64, max int64) int64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func minAmount(a Amount, b Amount) Amount {
	if a < b {
		return a
	}
	return b
}

func maxAmount(a Amount, b Amount) Amount {
	if a > b {
		return a
	}
	return b
}

func sumAmounts(values map[string]Amount) Amount {
	total := Amount(0)
	for _, value := range values {
		total += value
	}
	return total
}

func copyAmountMap(input map[string]Amount) map[string]Amount {
	out := make(map[string]Amount, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func sortedAmountBuckets(input map[string]Amount) []AmountBucket {
	keys := sortedKeys(input)
	out := make([]AmountBucket, 0, len(keys))
	for _, key := range keys {
		out = append(out, AmountBucket{Asset: key, Amount: input[key]})
	}
	return out
}

func amountFromSigned(value int64) (Amount, bool) {
	if value < 0 {
		return Amount(-value), true
	}
	return Amount(value), false
}

const maxInt64 = int64(^uint64(0) >> 1)

type AmountBucket struct {
	Asset  string `json:"asset"`
	Amount Amount `json:"amount"`
}

type SignedBucket struct {
	Asset  string `json:"asset"`
	Amount int64  `json:"amount"`
}

func signedBucket(asset string, amount int64) SignedBucket {
	return SignedBucket{Asset: asset, Amount: amount}
}
