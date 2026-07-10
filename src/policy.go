package main

type AdmissionPolicy struct {
	Name                   string `json:"name"`
	MaxRouteUtilizationBps int64  `json:"max_route_utilization_bps"`
	MinAccountEquityBps    int64  `json:"min_account_equity_bps"`
	MaxOpenPositions       int    `json:"max_open_positions"`
	AllowRouteRotation     bool   `json:"allow_route_rotation"`
	RequireFreshSnapshot   bool   `json:"require_fresh_snapshot"`
}

func DefaultAdmissionPolicy() AdmissionPolicy {
	return AdmissionPolicy{
		Name:                   "default",
		MaxRouteUtilizationBps: 9_600,
		MinAccountEquityBps:    1_200,
		MaxOpenPositions:       8,
		AllowRouteRotation:     true,
		RequireFreshSnapshot:   true,
	}
}

type AdmissionDecision struct {
	Allowed bool     `json:"allowed"`
	Reasons []string `json:"reasons"`
}

func (policy AdmissionPolicy) EvaluateOpen(account *Account, route *Route, notional Amount, margin Amount) AdmissionDecision {
	decision := AdmissionDecision{Allowed: true, Reasons: make([]string, 0)}
	if account.Status != AccountActive {
		decision.reject("account is not active")
	}
	if route.Status != RouteOpen {
		decision.reject("route is not open")
	}
	projected := route.Utilized + route.Reserved + notional
	projectedBps := int64(projected) * BpsScale / int64(route.Liquidity)
	if projectedBps > policy.MaxRouteUtilizationBps {
		decision.reject("projected route utilization exceeds policy")
	}
	if account.OpenPositionCount() >= policy.MaxOpenPositions {
		decision.reject("account has too many open positions")
	}
	if notional > 0 {
		marginBps := int64(margin) * BpsScale / int64(notional)
		if marginBps < policy.MinAccountEquityBps {
			decision.reject("position margin below policy")
		}
	}
	return decision
}

func (policy AdmissionPolicy) EvaluateRotation(account *Account, route *Route) AdmissionDecision {
	decision := AdmissionDecision{Allowed: true, Reasons: make([]string, 0)}
	if !policy.AllowRouteRotation {
		decision.reject("route rotation disabled")
	}
	if account.Status != AccountActive {
		decision.reject("account is not active")
	}
	if route.Status != RouteOpen {
		decision.reject("target route is not open")
	}
	if policy.RequireFreshSnapshot && account.SnapshotEpoch < route.LastEpoch-1 {
		decision.reject("account snapshot is stale")
	}
	return decision
}

func (decision *AdmissionDecision) reject(reason string) {
	decision.Allowed = false
	decision.Reasons = append(decision.Reasons, reason)
}

type StressStep struct {
	Label      string  `json:"label"`
	RouteID    RouteID `json:"route_id"`
	Notional   Amount  `json:"notional"`
	Epochs     int     `json:"epochs"`
	RotateTo   RouteID `json:"rotate_to"`
	CloseAfter bool    `json:"close_after"`
	Liquidate  bool    `json:"liquidate"`
}

type StressPlan struct {
	Name  string       `json:"name"`
	Steps []StressStep `json:"steps"`
}

func NewStressPlan(name string) StressPlan {
	return StressPlan{Name: name, Steps: make([]StressStep, 0)}
}

func (plan StressPlan) WithStep(step StressStep) StressPlan {
	plan.Steps = append(plan.Steps, step)
	return plan
}

func BaselineStressPlan() StressPlan {
	return NewStressPlan("baseline").
		WithStep(StressStep{Label: "atlas high utilization", RouteID: NewRouteID("atlas-eu"), Notional: MustAmount(820_000), Epochs: 2}).
		WithStep(StressStep{Label: "boreal medium utilization", RouteID: NewRouteID("boreal-us"), Notional: MustAmount(460_000), Epochs: 1}).
		WithStep(StressStep{Label: "cirrus low utilization", RouteID: NewRouteID("cirrus-apac"), Notional: MustAmount(180_000), Epochs: 1})
}
