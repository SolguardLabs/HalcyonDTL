package main

type UtilizationObservation struct {
	Epoch          EpochID `json:"epoch"`
	RouteID        RouteID `json:"route_id"`
	UtilizationBps int64   `json:"utilization_bps"`
	RatePPM        int64   `json:"rate_ppm"`
	Accumulator    int64   `json:"accumulator"`
	WeightBps      int64   `json:"weight_bps"`
	Source         string  `json:"source"`
}

type RouteOracleWindow struct {
	RouteID         RouteID                  `json:"route_id"`
	Observations    []UtilizationObservation `json:"observations"`
	WeightedUtilBps int64                    `json:"weighted_util_bps"`
	WeightedRatePPM int64                    `json:"weighted_rate_ppm"`
	LastAccumulator int64                    `json:"last_accumulator"`
	LastEpoch       EpochID                  `json:"last_epoch"`
	MaxWindow       int                      `json:"max_window"`
}

type UtilizationOracle struct {
	Windows   map[RouteID]*RouteOracleWindow `json:"windows"`
	MaxWindow int                            `json:"max_window"`
}

func NewUtilizationOracle(maxWindow int) *UtilizationOracle {
	if maxWindow <= 0 {
		maxWindow = 8
	}
	return &UtilizationOracle{
		Windows:   make(map[RouteID]*RouteOracleWindow),
		MaxWindow: maxWindow,
	}
}

func (oracle *UtilizationOracle) Observe(route *Route, epoch EpochID, source string) UtilizationObservation {
	window := oracle.window(route.ID)
	observation := UtilizationObservation{
		Epoch:          epoch,
		RouteID:        route.ID,
		UtilizationBps: route.UtilizationBps(),
		RatePPM:        route.LastFundingRatePPM,
		Accumulator:    route.FundingAccumulator,
		WeightBps:      BpsScale,
		Source:         nonEmpty(source, "engine"),
	}
	window.Append(observation)
	return observation
}

func (oracle *UtilizationOracle) ObserveAll(engine *Engine, source string) []UtilizationObservation {
	out := make([]UtilizationObservation, 0, len(engine.Routes))
	for _, routeID := range sortedRouteIDs(engine.Routes) {
		out = append(out, oracle.Observe(engine.Routes[routeID], engine.Clock, source))
	}
	return out
}

func (oracle *UtilizationOracle) window(routeID RouteID) *RouteOracleWindow {
	window, ok := oracle.Windows[routeID]
	if ok {
		return window
	}
	window = &RouteOracleWindow{
		RouteID:      routeID,
		Observations: make([]UtilizationObservation, 0, oracle.MaxWindow),
		MaxWindow:    oracle.MaxWindow,
	}
	oracle.Windows[routeID] = window
	return window
}

func (window *RouteOracleWindow) Append(observation UtilizationObservation) {
	window.Observations = append(window.Observations, observation)
	if len(window.Observations) > window.MaxWindow {
		window.Observations = window.Observations[len(window.Observations)-window.MaxWindow:]
	}
	window.LastAccumulator = observation.Accumulator
	window.LastEpoch = observation.Epoch
	window.recompute()
}

func (window *RouteOracleWindow) recompute() {
	if len(window.Observations) == 0 {
		window.WeightedUtilBps = 0
		window.WeightedRatePPM = 0
		return
	}
	var totalWeight int64
	var weightedUtil int64
	var weightedRate int64
	for index, observation := range window.Observations {
		ageWeight := int64(index + 1)
		weight := observation.WeightBps * ageWeight
		totalWeight += weight
		weightedUtil += observation.UtilizationBps * weight
		weightedRate += observation.RatePPM * weight
	}
	if totalWeight == 0 {
		window.WeightedUtilBps = 0
		window.WeightedRatePPM = 0
		return
	}
	window.WeightedUtilBps = weightedUtil / totalWeight
	window.WeightedRatePPM = weightedRate / totalWeight
}

func (window *RouteOracleWindow) Last() (UtilizationObservation, bool) {
	if len(window.Observations) == 0 {
		return UtilizationObservation{}, false
	}
	return window.Observations[len(window.Observations)-1], true
}

func (window *RouteOracleWindow) DriftFromRoute(route *Route) int64 {
	return route.FundingAccumulator - window.LastAccumulator
}

func (window *RouteOracleWindow) IsStale(current EpochID, maxAge int64) bool {
	if len(window.Observations) == 0 {
		return true
	}
	return int64(current-window.LastEpoch) > maxAge
}

func (oracle *UtilizationOracle) Reports() []OracleWindowReport {
	out := make([]OracleWindowReport, 0, len(oracle.Windows))
	keys := make(map[RouteID]*Route, len(oracle.Windows))
	for routeID := range oracle.Windows {
		keys[routeID] = nil
	}
	for _, routeID := range sortedRouteIDs(keys) {
		window := oracle.Windows[routeID]
		out = append(out, OracleWindowReport{
			RouteID:          window.RouteID.String(),
			WeightedUtilBps:  window.WeightedUtilBps,
			WeightedRatePPM:  window.WeightedRatePPM,
			LastAccumulator:  window.LastAccumulator,
			LastEpoch:        window.LastEpoch,
			ObservationCount: len(window.Observations),
			MaxWindow:        window.MaxWindow,
		})
	}
	return out
}

type OracleWindowReport struct {
	RouteID          string  `json:"route_id"`
	WeightedUtilBps  int64   `json:"weighted_util_bps"`
	WeightedRatePPM  int64   `json:"weighted_rate_ppm"`
	LastAccumulator  int64   `json:"last_accumulator"`
	LastEpoch        EpochID `json:"last_epoch"`
	ObservationCount int     `json:"observation_count"`
	MaxWindow        int     `json:"max_window"`
}
