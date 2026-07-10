package main

type LiquidationCheck struct {
	PositionID        PositionID `json:"position_id"`
	AccountID         AccountID  `json:"account_id"`
	RouteID           RouteID    `json:"route_id"`
	Asset             string     `json:"asset"`
	Notional          Amount     `json:"notional"`
	Equity            int64      `json:"equity"`
	MarginRatioBps    int64      `json:"margin_ratio_bps"`
	MaintenanceBps    int64      `json:"maintenance_bps"`
	UnrealizedFunding int64      `json:"unrealized_funding"`
	Liquidatable      bool       `json:"liquidatable"`
}

func (engine *Engine) CheckLiquidation(positionID PositionID) (LiquidationCheck, error) {
	position, err := engine.Position(positionID)
	if err != nil {
		return LiquidationCheck{}, err
	}
	route, err := engine.Route(position.RouteID)
	if err != nil {
		return LiquidationCheck{}, err
	}
	asset, err := engine.Asset(position.Asset)
	if err != nil {
		return LiquidationCheck{}, err
	}
	ratio := position.MarginRatioBps(route)
	check := LiquidationCheck{
		PositionID:        position.ID,
		AccountID:         position.AccountID,
		RouteID:           position.RouteID,
		Asset:             position.Asset,
		Notional:          position.Notional,
		Equity:            position.EquityWithRoute(route),
		MarginRatioBps:    ratio,
		MaintenanceBps:    asset.MaintenanceMarginBps,
		UnrealizedFunding: position.UnrealizedFunding(route),
		Liquidatable:      ratio < asset.MaintenanceMarginBps,
	}
	return check, nil
}

func (engine *Engine) LiquidatePosition(positionID PositionID, liquidator AccountID) (FundingSettlement, error) {
	check, err := engine.CheckLiquidation(positionID)
	if err != nil {
		return FundingSettlement{}, err
	}
	if !check.Liquidatable {
		return FundingSettlement{}, fail("engine.liquidate", "position %s is not liquidatable", positionID)
	}
	account, err := engine.Account(check.AccountID)
	if err != nil {
		return FundingSettlement{}, err
	}
	position, err := engine.Position(positionID)
	if err != nil {
		return FundingSettlement{}, err
	}
	route, err := engine.Route(position.RouteID)
	if err != nil {
		return FundingSettlement{}, err
	}
	operator, err := engine.Operator(route.Operator)
	if err != nil {
		return FundingSettlement{}, err
	}
	asset, err := engine.Asset(position.Asset)
	if err != nil {
		return FundingSettlement{}, err
	}
	settlement := SettlementUsesAccountSnapshot(account, position, route)
	expected := SettlementUsesRouteAccumulator(account, position, route)
	closeFee, err := operator.FeeFor(position.Notional)
	if err != nil {
		return FundingSettlement{}, err
	}
	liquidationFee, err := asset.LiquidationPenalty(position.Notional)
	if err != nil {
		return FundingSettlement{}, err
	}
	socialized, err := engine.applyPositionSettlement(account, position, route, operator, asset, settlement, closeFee, liquidationFee)
	if err != nil {
		return FundingSettlement{}, err
	}
	settlement.SocializedDebt = socialized
	if expected.Funding < settlement.Funding {
		engine.Pool.RecordUncollectedFunding(position.Asset, Amount(settlement.Funding-expected.Funding))
	}
	position.MarkLiquidated(engine.Clock, settlement, closeFee, liquidationFee)
	account.RemovePosition(position.ID)
	account.Status = AccountRestricted
	route.Release(position.Notional)
	if rewardAccount, ok := engine.Accounts[liquidator]; ok && liquidationFee > 0 {
		reward := liquidationFee / 2
		rewardAccount.Credit(position.Asset, reward)
	}
	engine.Journal.Append(engine.Clock, "position.liquidated", eventFields(
		"account", account.ID.String(),
		"liquidator", liquidator.String(),
		"route", route.ID.String(),
		"position", position.ID.String(),
		"funding", settlement.Funding,
		"expected_funding", expected.Funding,
		"close_fee", closeFee,
		"liquidation_fee", liquidationFee,
		"socialized_debt", socialized,
	))
	return settlement, nil
}

func (engine *Engine) LiquidatablePositions() []LiquidationCheck {
	out := make([]LiquidationCheck, 0)
	for _, positionID := range sortedPositionIDs(engine.Positions) {
		position := engine.Positions[positionID]
		if !position.IsOpen() {
			continue
		}
		check, err := engine.CheckLiquidation(positionID)
		if err == nil && check.Liquidatable {
			out = append(out, check)
		}
	}
	return out
}
