package main

type HalcyonReport struct {
	Lab          string             `json:"lab"`
	Scenario     string             `json:"scenario"`
	NetworkID    string             `json:"network_id"`
	Clock        EpochID            `json:"clock"`
	StateDigest  string             `json:"state_digest"`
	Assets       []Asset            `json:"assets"`
	Operators    []OperatorReport   `json:"operators"`
	Vaults       []VaultReport      `json:"vaults"`
	Accounts     []AccountReport    `json:"accounts"`
	Routes       []RouteReport      `json:"routes"`
	Positions    []PositionReport   `json:"positions"`
	Funding      FundingReport      `json:"funding"`
	Pool         PoolReport         `json:"pool"`
	RouteDrift   []RouteDrift       `json:"route_drift"`
	Liquidations []LiquidationCheck `json:"liquidations"`
	Totals       TotalsReport       `json:"totals"`
	Risk         RiskReport         `json:"risk"`
	Invariants   map[string]bool    `json:"invariants"`
	Events       []Event            `json:"events"`
	Notes        []string           `json:"notes"`
}

type OperatorReport struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	FeeBps          int64          `json:"fee_bps"`
	FundingShareBps int64          `json:"funding_share_bps"`
	RiskWeightBps   int64          `json:"risk_weight_bps"`
	Balances        []AmountBucket `json:"balances"`
	RouteCount      int            `json:"route_count"`
	Status          string         `json:"status"`
}

type VaultReport struct {
	ID             string  `json:"id"`
	Asset          string  `json:"asset"`
	Operator       string  `json:"operator"`
	Reserve        Amount  `json:"reserve"`
	LockedMargin   Amount  `json:"locked_margin"`
	Insurance      Amount  `json:"insurance"`
	PendingRebate  Amount  `json:"pending_rebate"`
	SocializedDebt Amount  `json:"socialized_debt"`
	MinReserve     Amount  `json:"min_reserve"`
	Available      Amount  `json:"available"`
	HealthBps      int64   `json:"health_bps"`
	Status         string  `json:"status"`
	LastAccounting EpochID `json:"last_accounting"`
}

type AccountReport struct {
	ID              string         `json:"id"`
	Owner           string         `json:"owner"`
	Collateral      []AmountBucket `json:"collateral"`
	ReservedMargin  []AmountBucket `json:"reserved_margin"`
	RealizedFunding []SignedBucket `json:"realized_funding"`
	FeesPaid        []AmountBucket `json:"fees_paid"`
	Positions       []string       `json:"positions"`
	ActiveRoute     string         `json:"active_route"`
	FundingSnapshot int64          `json:"funding_snapshot"`
	SnapshotEpoch   EpochID        `json:"snapshot_epoch"`
	SnapshotRoute   string         `json:"snapshot_route"`
	Status          string         `json:"status"`
	Warnings        []string       `json:"warnings"`
}

type RouteReport struct {
	ID                   string          `json:"id"`
	SourceVault          string          `json:"source_vault"`
	DestinationVault     string          `json:"destination_vault"`
	Operator             string          `json:"operator"`
	Asset                string          `json:"asset"`
	Liquidity            Amount          `json:"liquidity"`
	Utilized             Amount          `json:"utilized"`
	Reserved             Amount          `json:"reserved"`
	Capacity             Amount          `json:"capacity"`
	UtilizationBps       int64           `json:"utilization_bps"`
	TargetUtilizationBps int64           `json:"target_utilization_bps"`
	MaxUtilizationBps    int64           `json:"max_utilization_bps"`
	LastFundingRatePPM   int64           `json:"last_funding_rate_ppm"`
	FundingAccumulator   int64           `json:"funding_accumulator"`
	LastEpoch            EpochID         `json:"last_epoch"`
	OpenInterest         int             `json:"open_interest"`
	Status               string          `json:"status"`
	EpochHistory         []FundingRecord `json:"epoch_history"`
	Warnings             []string        `json:"warnings"`
}

type PositionReport struct {
	ID                string  `json:"id"`
	AccountID         string  `json:"account_id"`
	RouteID           string  `json:"route_id"`
	Asset             string  `json:"asset"`
	Direction         string  `json:"direction"`
	Notional          Amount  `json:"notional"`
	Margin            Amount  `json:"margin"`
	EntryAccumulator  int64   `json:"entry_accumulator"`
	ExitAccumulator   int64   `json:"exit_accumulator"`
	OpenEpoch         EpochID `json:"open_epoch"`
	CloseEpoch        EpochID `json:"close_epoch"`
	FundingPaid       int64   `json:"funding_paid"`
	CloseFee          Amount  `json:"close_fee"`
	LiquidationFee    Amount  `json:"liquidation_fee"`
	SocializedDebt    Amount  `json:"socialized_debt"`
	Status            string  `json:"status"`
	CloseReason       string  `json:"close_reason"`
	UnrealizedFunding int64   `json:"unrealized_funding"`
	MarginRatioBps    int64   `json:"margin_ratio_bps"`
}

type FundingReport struct {
	Records          []FundingRecord `json:"records"`
	LastEpoch        EpochID         `json:"last_epoch"`
	NegativeNotional Amount          `json:"negative_notional"`
	PositiveNotional Amount          `json:"positive_notional"`
}

type PoolReport struct {
	FeesCollected      []AmountBucket `json:"fees_collected"`
	InsuranceBalance   []AmountBucket `json:"insurance_balance"`
	SocializedDebt     []AmountBucket `json:"socialized_debt"`
	UncollectedFunding []AmountBucket `json:"uncollected_funding"`
}

type RouteDrift struct {
	AccountID        string  `json:"account_id"`
	RouteID          string  `json:"route_id"`
	ActiveRoute      string  `json:"active_route"`
	Notional         Amount  `json:"notional"`
	AccountSnapshot  int64   `json:"account_snapshot"`
	RouteAccumulator int64   `json:"route_accumulator"`
	AccumulatorDelta int64   `json:"accumulator_delta"`
	SnapshotRoute    string  `json:"snapshot_route"`
	SnapshotEpoch    EpochID `json:"snapshot_epoch"`
	UsesActiveCursor bool    `json:"uses_active_cursor"`
}

type TotalsReport struct {
	Collateral         []AmountBucket `json:"collateral"`
	ReservedMargin     []AmountBucket `json:"reserved_margin"`
	VaultReserves      []AmountBucket `json:"vault_reserves"`
	RouteLiquidity     []AmountBucket `json:"route_liquidity"`
	RouteUtilized      []AmountBucket `json:"route_utilized"`
	FeesCollected      []AmountBucket `json:"fees_collected"`
	SocializedDebt     []AmountBucket `json:"socialized_debt"`
	UncollectedFunding []AmountBucket `json:"uncollected_funding"`
}

type RiskReport struct {
	OpenPositions         int    `json:"open_positions"`
	ClosedPositions       int    `json:"closed_positions"`
	LiquidatedPositions   int    `json:"liquidated_positions"`
	MaxUtilizationBps     int64  `json:"max_utilization_bps"`
	MinMarginRatioBps     int64  `json:"min_margin_ratio_bps"`
	ProtocolFeesCollected Amount `json:"protocol_fees_collected"`
	InsuranceBalance      Amount `json:"insurance_balance"`
	SocializedDebt        Amount `json:"socialized_debt"`
	UncollectedFunding    Amount `json:"uncollected_funding"`
	FundingRecords        int    `json:"funding_records"`
	EventCount            int    `json:"event_count"`
}

func (engine *Engine) Report(scenario string) HalcyonReport {
	risk := engine.ComputeRisk()
	return HalcyonReport{
		Lab:          "HalcyonDTL",
		Scenario:     scenario,
		NetworkID:    engine.NetworkID,
		Clock:        engine.Clock,
		StateDigest:  engine.StateDigest(),
		Assets:       engine.Assets.List(),
		Operators:    engine.OperatorReports(),
		Vaults:       engine.VaultReports(),
		Accounts:     engine.AccountReports(),
		Routes:       engine.RouteReports(),
		Positions:    engine.PositionReports(),
		Funding:      engine.Funding.Report(),
		Pool:         engine.Pool.Report(),
		RouteDrift:   engine.RouteDriftReports(),
		Liquidations: engine.LiquidatablePositions(),
		Totals:       engine.Totals(),
		Risk:         risk,
		Invariants:   engine.Invariants(),
		Events:       engine.Journal.Events(),
		Notes:        cloneStrings(engine.Notes),
	}
}

func (engine *Engine) OperatorReports() []OperatorReport {
	out := make([]OperatorReport, 0, len(engine.Operators))
	for _, id := range sortedOperatorIDs(engine.Operators) {
		out = append(out, engine.Operators[id].Report())
	}
	return out
}

func (engine *Engine) VaultReports() []VaultReport {
	out := make([]VaultReport, 0, len(engine.Vaults))
	for _, id := range sortedVaultIDs(engine.Vaults) {
		out = append(out, engine.Vaults[id].Report())
	}
	return out
}

func (engine *Engine) AccountReports() []AccountReport {
	out := make([]AccountReport, 0, len(engine.Accounts))
	for _, id := range sortedAccountIDs(engine.Accounts) {
		out = append(out, engine.Accounts[id].Report())
	}
	return out
}

func (engine *Engine) RouteReports() []RouteReport {
	out := make([]RouteReport, 0, len(engine.Routes))
	for _, id := range sortedRouteIDs(engine.Routes) {
		out = append(out, engine.Routes[id].Report())
	}
	return out
}

func (engine *Engine) PositionReports() []PositionReport {
	out := make([]PositionReport, 0, len(engine.Positions))
	for _, id := range sortedPositionIDs(engine.Positions) {
		position := engine.Positions[id]
		route := engine.Routes[position.RouteID]
		out = append(out, position.Report(route))
	}
	return out
}

func (engine *Engine) RouteDriftReports() []RouteDrift {
	out := make([]RouteDrift, 0)
	for _, accountID := range sortedAccountIDs(engine.Accounts) {
		out = append(out, engine.AccountRouteDrift(accountID)...)
	}
	return out
}

func (engine *Engine) Totals() TotalsReport {
	collateral := make(map[string]Amount)
	reserved := make(map[string]Amount)
	vaults := make(map[string]Amount)
	routeLiquidity := make(map[string]Amount)
	routeUtilized := make(map[string]Amount)
	for _, account := range engine.Accounts {
		for asset, amount := range account.Collateral {
			collateral[asset] += amount
		}
		for asset, amount := range account.ReservedMargin {
			reserved[asset] += amount
		}
	}
	for _, vault := range engine.Vaults {
		vaults[vault.Asset] += vault.Reserve
	}
	for _, route := range engine.Routes {
		routeLiquidity[route.Asset] += route.Liquidity
		routeUtilized[route.Asset] += route.Utilized
	}
	return TotalsReport{
		Collateral:         sortedAmountBuckets(collateral),
		ReservedMargin:     sortedAmountBuckets(reserved),
		VaultReserves:      sortedAmountBuckets(vaults),
		RouteLiquidity:     sortedAmountBuckets(routeLiquidity),
		RouteUtilized:      sortedAmountBuckets(routeUtilized),
		FeesCollected:      sortedAmountBuckets(engine.Pool.FeesCollected),
		SocializedDebt:     sortedAmountBuckets(engine.Pool.SocializedDebt),
		UncollectedFunding: sortedAmountBuckets(engine.Pool.UncollectedFunding),
	}
}
