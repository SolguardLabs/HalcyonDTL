package main

func (engine *Engine) ComputeRisk() RiskReport {
	report := RiskReport{
		MinMarginRatioBps: BpsScale,
		FundingRecords:    len(engine.Funding.Records),
		EventCount:        len(engine.Journal.Events()),
	}
	for _, route := range engine.Routes {
		if route.UtilizationBps() > report.MaxUtilizationBps {
			report.MaxUtilizationBps = route.UtilizationBps()
		}
	}
	for _, position := range engine.Positions {
		switch position.Status {
		case PositionOpen, PositionClosing:
			report.OpenPositions++
			if route := engine.Routes[position.RouteID]; route != nil {
				ratio := position.MarginRatioBps(route)
				if ratio < report.MinMarginRatioBps {
					report.MinMarginRatioBps = ratio
				}
			}
		case PositionClosed:
			report.ClosedPositions++
		case PositionLiquidated:
			report.LiquidatedPositions++
		}
	}
	for _, amount := range engine.Pool.FeesCollected {
		report.ProtocolFeesCollected += amount
	}
	for _, amount := range engine.Pool.InsuranceBalance {
		report.InsuranceBalance += amount
	}
	for _, amount := range engine.Pool.SocializedDebt {
		report.SocializedDebt += amount
	}
	for _, amount := range engine.Pool.UncollectedFunding {
		report.UncollectedFunding += amount
	}
	return report
}

func (engine *Engine) Invariants() map[string]bool {
	return map[string]bool{
		"accounts_non_negative":             engine.accountsNonNegative(),
		"vaults_non_negative":               engine.vaultsNonNegative(),
		"routes_within_max_utilization":     engine.routesWithinMaxUtilization(),
		"positions_linked":                  engine.positionsLinked(),
		"funding_records_link_routes":       engine.fundingRecordsLinkRoutes(),
		"active_snapshots_link_open_routes": engine.activeSnapshotsLinkOpenRoutes(),
		"closed_positions_not_in_accounts":  engine.closedPositionsNotInAccounts(),
	}
}

func (engine *Engine) Validate() []string {
	failures := make([]string, 0)
	for name, ok := range engine.Invariants() {
		if !ok {
			failures = append(failures, name)
		}
	}
	sortStrings(failures)
	return failures
}

func (engine *Engine) accountsNonNegative() bool {
	for _, account := range engine.Accounts {
		for _, amount := range account.Collateral {
			if amount < 0 {
				return false
			}
		}
		for _, amount := range account.ReservedMargin {
			if amount < 0 {
				return false
			}
		}
	}
	return true
}

func (engine *Engine) vaultsNonNegative() bool {
	for _, vault := range engine.Vaults {
		if vault.Reserve < 0 || vault.LockedMargin < 0 || vault.Insurance < 0 || vault.SocializedDebt < 0 {
			return false
		}
	}
	return true
}

func (engine *Engine) routesWithinMaxUtilization() bool {
	for _, route := range engine.Routes {
		if route.UtilizationBps() > route.MaxUtilizationBps {
			return false
		}
	}
	return true
}

func (engine *Engine) positionsLinked() bool {
	for _, position := range engine.Positions {
		if _, ok := engine.Accounts[position.AccountID]; !ok {
			return false
		}
		if _, ok := engine.Routes[position.RouteID]; !ok {
			return false
		}
	}
	return true
}

func (engine *Engine) fundingRecordsLinkRoutes() bool {
	for _, record := range engine.Funding.Records {
		if _, ok := engine.Routes[RouteID(record.RouteID)]; !ok {
			return false
		}
	}
	return true
}

func (engine *Engine) activeSnapshotsLinkOpenRoutes() bool {
	for _, account := range engine.Accounts {
		if account.ActiveRoute.Empty() {
			continue
		}
		route, ok := engine.Routes[account.ActiveRoute]
		if !ok {
			return false
		}
		if route.Status == RouteClosed {
			return false
		}
	}
	return true
}

func (engine *Engine) closedPositionsNotInAccounts() bool {
	for _, account := range engine.Accounts {
		for positionID := range account.Positions {
			position := engine.Positions[positionID]
			if position == nil || !position.IsOpen() {
				return false
			}
		}
	}
	return true
}
