package main

type RouteRotation struct {
	AccountID        AccountID `json:"account_id"`
	PreviousRoute    RouteID   `json:"previous_route"`
	NextRoute        RouteID   `json:"next_route"`
	Epoch            EpochID   `json:"epoch"`
	PreviousSnapshot int64     `json:"previous_snapshot"`
	NextSnapshot     int64     `json:"next_snapshot"`
	Reason           string    `json:"reason"`
}

func (engine *Engine) PreviewRotation(accountID AccountID, routeID RouteID, reason string) (RouteRotation, error) {
	account, err := engine.Account(accountID)
	if err != nil {
		return RouteRotation{}, err
	}
	route, err := engine.Route(routeID)
	if err != nil {
		return RouteRotation{}, err
	}
	return RouteRotation{
		AccountID:        account.ID,
		PreviousRoute:    account.ActiveRoute,
		NextRoute:        route.ID,
		Epoch:            engine.Clock,
		PreviousSnapshot: account.FundingSnapshot,
		NextSnapshot:     route.FundingAccumulator,
		Reason:           reason,
	}, nil
}

func (engine *Engine) ApplyRotation(accountID AccountID, routeID RouteID, reason string) (RouteRotation, error) {
	preview, err := engine.PreviewRotation(accountID, routeID, reason)
	if err != nil {
		return RouteRotation{}, err
	}
	if err := engine.RotateAccountRoute(accountID, routeID, reason); err != nil {
		return RouteRotation{}, err
	}
	return preview, nil
}

func (engine *Engine) AccountExposureByRoute(accountID AccountID) map[RouteID]Amount {
	out := make(map[RouteID]Amount)
	account, ok := engine.Accounts[accountID]
	if !ok {
		return out
	}
	for positionID := range account.Positions {
		position := engine.Positions[positionID]
		if position == nil || !position.IsOpen() {
			continue
		}
		out[position.RouteID] += position.Notional
	}
	return out
}

func (engine *Engine) AccountRouteDrift(accountID AccountID) []RouteDrift {
	account, ok := engine.Accounts[accountID]
	if !ok {
		return nil
	}
	exposure := engine.AccountExposureByRoute(accountID)
	out := make([]RouteDrift, 0, len(exposure))
	for routeID, notional := range exposure {
		route := engine.Routes[routeID]
		if route == nil {
			continue
		}
		out = append(out, RouteDrift{
			AccountID:        account.ID.String(),
			RouteID:          route.ID.String(),
			ActiveRoute:      account.ActiveRoute.String(),
			Notional:         notional,
			AccountSnapshot:  account.FundingSnapshot,
			RouteAccumulator: route.FundingAccumulator,
			AccumulatorDelta: route.FundingAccumulator - account.FundingSnapshot,
			SnapshotRoute:    account.SnapshotRoute.String(),
			SnapshotEpoch:    account.SnapshotEpoch,
			UsesActiveCursor: account.ActiveRoute == route.ID,
		})
	}
	return out
}
