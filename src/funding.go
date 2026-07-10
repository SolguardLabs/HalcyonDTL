package main

type FundingRecord struct {
	Epoch             EpochID `json:"epoch"`
	RouteID           string  `json:"route_id"`
	Asset             string  `json:"asset"`
	UtilizationBps    int64   `json:"utilization_bps"`
	TargetBps         int64   `json:"target_bps"`
	RatePPM           int64   `json:"rate_ppm"`
	AccumulatorBefore int64   `json:"accumulator_before"`
	AccumulatorAfter  int64   `json:"accumulator_after"`
	OpenInterest      int     `json:"open_interest"`
	Utilized          Amount  `json:"utilized"`
	Liquidity         Amount  `json:"liquidity"`
}

type FundingSettlement struct {
	AccountID        AccountID  `json:"account_id"`
	PositionID       PositionID `json:"position_id"`
	RouteID          RouteID    `json:"route_id"`
	Asset            string     `json:"asset"`
	Notional         Amount     `json:"notional"`
	EntryAccumulator int64      `json:"entry_accumulator"`
	ExitAccumulator  int64      `json:"exit_accumulator"`
	RouteAccumulator int64      `json:"route_accumulator"`
	DeltaPPM         int64      `json:"delta_ppm"`
	Funding          int64      `json:"funding"`
	SocializedDebt   Amount     `json:"socialized_debt"`
	Reason           string     `json:"reason"`
}

type FundingEngine struct {
	Records          []FundingRecord `json:"records"`
	LastEpoch        EpochID         `json:"last_epoch"`
	NegativeNotional Amount          `json:"negative_notional"`
	PositiveNotional Amount          `json:"positive_notional"`
}

func NewFundingEngine() *FundingEngine {
	return &FundingEngine{Records: make([]FundingRecord, 0)}
}

func (funding *FundingEngine) Append(record FundingRecord) {
	funding.Records = append(funding.Records, record)
	funding.LastEpoch = record.Epoch
	if record.RatePPM < 0 {
		funding.NegativeNotional += record.Utilized
	} else if record.RatePPM > 0 {
		funding.PositiveNotional += record.Utilized
	}
}

func (funding *FundingEngine) Report() FundingReport {
	records := make([]FundingRecord, len(funding.Records))
	copy(records, funding.Records)
	return FundingReport{
		Records:          records,
		LastEpoch:        funding.LastEpoch,
		NegativeNotional: funding.NegativeNotional,
		PositiveNotional: funding.PositiveNotional,
	}
}

func FundingDebitAmount(value int64) Amount {
	if value >= 0 {
		return 0
	}
	return Amount(-value)
}

func FundingCreditAmount(value int64) Amount {
	if value <= 0 {
		return 0
	}
	return Amount(value)
}

func SettlementUsesAccountSnapshot(account *Account, position *Position, route *Route) FundingSettlement {
	exitAccumulator := account.FundingSnapshot
	deltaPPM := exitAccumulator - position.EntryAccumulator
	funding, _ := SignedAmountFromPPM(position.Notional, deltaPPM)
	return FundingSettlement{
		AccountID:        account.ID,
		PositionID:       position.ID,
		RouteID:          position.RouteID,
		Asset:            position.Asset,
		Notional:         position.Notional,
		EntryAccumulator: position.EntryAccumulator,
		ExitAccumulator:  exitAccumulator,
		RouteAccumulator: route.FundingAccumulator,
		DeltaPPM:         deltaPPM,
		Funding:          funding,
		Reason:           "account_snapshot",
	}
}

func SettlementUsesRouteAccumulator(account *Account, position *Position, route *Route) FundingSettlement {
	exitAccumulator := route.FundingAccumulator
	deltaPPM := exitAccumulator - position.EntryAccumulator
	funding, _ := SignedAmountFromPPM(position.Notional, deltaPPM)
	return FundingSettlement{
		AccountID:        account.ID,
		PositionID:       position.ID,
		RouteID:          position.RouteID,
		Asset:            position.Asset,
		Notional:         position.Notional,
		EntryAccumulator: position.EntryAccumulator,
		ExitAccumulator:  exitAccumulator,
		RouteAccumulator: route.FundingAccumulator,
		DeltaPPM:         deltaPPM,
		Funding:          funding,
		Reason:           "route_accumulator",
	}
}

func FundingGap(vulnerable FundingSettlement, expected FundingSettlement) int64 {
	return expected.Funding - vulnerable.Funding
}
