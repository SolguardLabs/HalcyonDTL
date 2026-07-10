package main

type PositionStatus string

const (
	PositionOpen       PositionStatus = "open"
	PositionClosing    PositionStatus = "closing"
	PositionClosed     PositionStatus = "closed"
	PositionLiquidated PositionStatus = "liquidated"
)

type Position struct {
	ID               PositionID     `json:"id"`
	AccountID        AccountID      `json:"account_id"`
	RouteID          RouteID        `json:"route_id"`
	Asset            string         `json:"asset"`
	Direction        string         `json:"direction"`
	Notional         Amount         `json:"notional"`
	Margin           Amount         `json:"margin"`
	EntryAccumulator int64          `json:"entry_accumulator"`
	ExitAccumulator  int64          `json:"exit_accumulator"`
	OpenEpoch        EpochID        `json:"open_epoch"`
	CloseEpoch       EpochID        `json:"close_epoch"`
	FundingPaid      int64          `json:"funding_paid"`
	CloseFee         Amount         `json:"close_fee"`
	LiquidationFee   Amount         `json:"liquidation_fee"`
	SocializedDebt   Amount         `json:"socialized_debt"`
	Status           PositionStatus `json:"status"`
	CloseReason      string         `json:"close_reason"`
}

func NewPosition(id PositionID, account AccountID, route *Route, notional Amount, margin Amount, epoch EpochID) *Position {
	return &Position{
		ID:               id,
		AccountID:        account,
		RouteID:          route.ID,
		Asset:            route.Asset,
		Direction:        "long-liquidity",
		Notional:         notional,
		Margin:           margin,
		EntryAccumulator: route.FundingAccumulator,
		OpenEpoch:        epoch,
		Status:           PositionOpen,
	}
}

func (position *Position) IsOpen() bool {
	return position.Status == PositionOpen || position.Status == PositionClosing
}

func (position *Position) Age(current EpochID) int64 {
	if current <= position.OpenEpoch {
		return 0
	}
	return int64(current - position.OpenEpoch)
}

func (position *Position) UnrealizedFunding(route *Route) int64 {
	funding, err := route.FundingFor(position.Notional, position.EntryAccumulator)
	if err != nil {
		return 0
	}
	return funding
}

func (position *Position) EquityWithRoute(route *Route) int64 {
	return int64(position.Margin) + position.UnrealizedFunding(route)
}

func (position *Position) MarginRatioBps(route *Route) int64 {
	if position.Notional == 0 {
		return BpsScale
	}
	return position.EquityWithRoute(route) * BpsScale / int64(position.Notional)
}

func (position *Position) MarkClosing() {
	if position.Status == PositionOpen {
		position.Status = PositionClosing
	}
}

func (position *Position) MarkClosed(epoch EpochID, settlement FundingSettlement, fee Amount, reason string) {
	position.Status = PositionClosed
	position.CloseEpoch = epoch
	position.ExitAccumulator = settlement.ExitAccumulator
	position.FundingPaid = settlement.Funding
	position.SocializedDebt = settlement.SocializedDebt
	position.CloseFee = fee
	position.CloseReason = reason
}

func (position *Position) MarkLiquidated(epoch EpochID, settlement FundingSettlement, closeFee Amount, liquidationFee Amount) {
	position.Status = PositionLiquidated
	position.CloseEpoch = epoch
	position.ExitAccumulator = settlement.ExitAccumulator
	position.FundingPaid = settlement.Funding
	position.SocializedDebt = settlement.SocializedDebt
	position.CloseFee = closeFee
	position.LiquidationFee = liquidationFee
	position.CloseReason = "liquidation"
}

func (position *Position) Report(route *Route) PositionReport {
	return PositionReport{
		ID:               position.ID.String(),
		AccountID:        position.AccountID.String(),
		RouteID:          position.RouteID.String(),
		Asset:            position.Asset,
		Direction:        position.Direction,
		Notional:         position.Notional,
		Margin:           position.Margin,
		EntryAccumulator: position.EntryAccumulator,
		ExitAccumulator:  position.ExitAccumulator,
		OpenEpoch:        position.OpenEpoch,
		CloseEpoch:       position.CloseEpoch,
		FundingPaid:      position.FundingPaid,
		CloseFee:         position.CloseFee,
		LiquidationFee:   position.LiquidationFee,
		SocializedDebt:   position.SocializedDebt,
		Status:           string(position.Status),
		CloseReason:      position.CloseReason,
		UnrealizedFunding: func() int64 {
			if route == nil || !position.IsOpen() {
				return 0
			}
			return position.UnrealizedFunding(route)
		}(),
		MarginRatioBps: func() int64 {
			if route == nil || !position.IsOpen() {
				return 0
			}
			return position.MarginRatioBps(route)
		}(),
	}
}
