package main

type RouteStatus string

const (
	RouteOpen     RouteStatus = "open"
	RoutePaused   RouteStatus = "paused"
	RouteSettling RouteStatus = "settling"
	RouteClosed   RouteStatus = "closed"
)

type RouteSpec struct {
	ID                   RouteID
	SourceVault          VaultID
	DestinationVault     VaultID
	Operator             OperatorID
	Asset                string
	Liquidity            Amount
	TargetUtilizationBps int64
	MaxUtilizationBps    int64
	FundingSlopePPM      int64
	FundingClampPPM      int64
}

type Route struct {
	ID                   RouteID         `json:"id"`
	SourceVault          VaultID         `json:"source_vault"`
	DestinationVault     VaultID         `json:"destination_vault"`
	Operator             OperatorID      `json:"operator"`
	Asset                string          `json:"asset"`
	Liquidity            Amount          `json:"liquidity"`
	Utilized             Amount          `json:"utilized"`
	Reserved             Amount          `json:"reserved"`
	TargetUtilizationBps int64           `json:"target_utilization_bps"`
	MaxUtilizationBps    int64           `json:"max_utilization_bps"`
	FundingSlopePPM      int64           `json:"funding_slope_ppm"`
	FundingClampPPM      int64           `json:"funding_clamp_ppm"`
	LastFundingRatePPM   int64           `json:"last_funding_rate_ppm"`
	FundingAccumulator   int64           `json:"funding_accumulator"`
	LastEpoch            EpochID         `json:"last_epoch"`
	OpenInterest         int             `json:"open_interest"`
	Status               RouteStatus     `json:"status"`
	EpochHistory         []FundingRecord `json:"epoch_history"`
	Warnings             []string        `json:"warnings"`
}

func NewRoute(spec RouteSpec) *Route {
	return &Route{
		ID:                   spec.ID,
		SourceVault:          spec.SourceVault,
		DestinationVault:     spec.DestinationVault,
		Operator:             spec.Operator,
		Asset:                spec.Asset,
		Liquidity:            spec.Liquidity,
		TargetUtilizationBps: spec.TargetUtilizationBps,
		MaxUtilizationBps:    spec.MaxUtilizationBps,
		FundingSlopePPM:      spec.FundingSlopePPM,
		FundingClampPPM:      spec.FundingClampPPM,
		Status:               RouteOpen,
		EpochHistory:         make([]FundingRecord, 0),
		Warnings:             make([]string, 0),
	}
}

func DefaultRouteSpec(label string, source VaultID, destination VaultID, operator OperatorID, asset Asset, liquidity Amount) RouteSpec {
	return RouteSpec{
		ID:                   NewRouteID(label),
		SourceVault:          source,
		DestinationVault:     destination,
		Operator:             operator,
		Asset:                asset.Symbol,
		Liquidity:            liquidity,
		TargetUtilizationBps: 6_800,
		MaxUtilizationBps:    9_600,
		FundingSlopePPM:      asset.FundingSlopePPM,
		FundingClampPPM:      asset.FundingClampPPM,
	}
}

func (route *Route) Validate() error {
	if route.ID.Empty() {
		return fail("route.validate", "missing route id")
	}
	if route.SourceVault.Empty() || route.DestinationVault.Empty() {
		return fail("route.validate", "route %s has empty vault endpoint", route.ID)
	}
	if route.Operator.Empty() {
		return fail("route.validate", "route %s has no operator", route.ID)
	}
	if route.Asset == "" {
		return fail("route.validate", "route %s has no asset", route.ID)
	}
	if route.Liquidity == 0 {
		return fail("route.validate", "route %s has no liquidity", route.ID)
	}
	if route.TargetUtilizationBps <= 0 || route.TargetUtilizationBps >= route.MaxUtilizationBps {
		return fail("route.validate", "route %s has invalid utilization target", route.ID)
	}
	if route.FundingClampPPM <= 0 || route.FundingSlopePPM <= 0 {
		return fail("route.validate", "route %s has invalid funding params", route.ID)
	}
	return nil
}

func (route *Route) Capacity() Amount {
	max, _ := route.Liquidity.MulBps(route.MaxUtilizationBps)
	if max <= route.Utilized+route.Reserved {
		return 0
	}
	return max - route.Utilized - route.Reserved
}

func (route *Route) UtilizationBps() int64 {
	if route.Liquidity == 0 {
		return 0
	}
	return int64(route.Utilized+route.Reserved) * BpsScale / int64(route.Liquidity)
}

func (route *Route) Reserve(notional Amount) error {
	if route.Status != RouteOpen {
		return fail("route.reserve", "route %s is not open", route.ID)
	}
	if route.Capacity() < notional {
		return fail("route.reserve", "route %s capacity exceeded", route.ID)
	}
	route.Reserved += notional
	return nil
}

func (route *Route) ActivateReservation(notional Amount) {
	if notional > route.Reserved {
		route.Reserved = 0
	} else {
		route.Reserved -= notional
	}
	route.Utilized += notional
	route.OpenInterest++
}

func (route *Route) Release(notional Amount) {
	if notional > route.Utilized {
		route.Utilized = 0
	} else {
		route.Utilized -= notional
	}
	if route.OpenInterest > 0 {
		route.OpenInterest--
	}
}

func (route *Route) Pause(reason string) {
	route.Status = RoutePaused
	route.Warnings = append(route.Warnings, reason)
}

func (route *Route) Resume() {
	if route.Status == RoutePaused {
		route.Status = RouteOpen
	}
}

func (route *Route) Close() {
	route.Status = RouteClosed
}

func (route *Route) FundingRatePPM() int64 {
	utilization := route.UtilizationBps()
	deviation := route.TargetUtilizationBps - utilization
	rate := deviation * route.FundingSlopePPM / BpsScale
	return clampInt64(rate, -route.FundingClampPPM, route.FundingClampPPM)
}

func (route *Route) ApplyFunding(epoch EpochID) FundingRecord {
	rate := route.FundingRatePPM()
	before := route.FundingAccumulator
	route.FundingAccumulator += rate
	route.LastFundingRatePPM = rate
	route.LastEpoch = epoch
	record := FundingRecord{
		Epoch:             epoch,
		RouteID:           route.ID.String(),
		Asset:             route.Asset,
		UtilizationBps:    route.UtilizationBps(),
		TargetBps:         route.TargetUtilizationBps,
		RatePPM:           rate,
		AccumulatorBefore: before,
		AccumulatorAfter:  route.FundingAccumulator,
		OpenInterest:      route.OpenInterest,
		Utilized:          route.Utilized,
		Liquidity:         route.Liquidity,
	}
	route.EpochHistory = append(route.EpochHistory, record)
	return record
}

func (route *Route) FundingFor(notional Amount, entryAccumulator int64) (int64, error) {
	delta := route.FundingAccumulator - entryAccumulator
	return SignedAmountFromPPM(notional, delta)
}

func (route *Route) Report() RouteReport {
	history := make([]FundingRecord, len(route.EpochHistory))
	copy(history, route.EpochHistory)
	return RouteReport{
		ID:                   route.ID.String(),
		SourceVault:          route.SourceVault.String(),
		DestinationVault:     route.DestinationVault.String(),
		Operator:             route.Operator.String(),
		Asset:                route.Asset,
		Liquidity:            route.Liquidity,
		Utilized:             route.Utilized,
		Reserved:             route.Reserved,
		Capacity:             route.Capacity(),
		UtilizationBps:       route.UtilizationBps(),
		TargetUtilizationBps: route.TargetUtilizationBps,
		MaxUtilizationBps:    route.MaxUtilizationBps,
		LastFundingRatePPM:   route.LastFundingRatePPM,
		FundingAccumulator:   route.FundingAccumulator,
		LastEpoch:            route.LastEpoch,
		OpenInterest:         route.OpenInterest,
		Status:               string(route.Status),
		EpochHistory:         history,
		Warnings:             cloneStrings(route.Warnings),
	}
}
